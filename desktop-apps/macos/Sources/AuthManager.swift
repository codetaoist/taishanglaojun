import Foundation
import Combine

// MARK: - 数据模型

struct User: Codable {
    let id: String
    let username: String
    let email: String
    let avatarUrl: String?
    let createdAt: String
    let updatedAt: String
    
    enum CodingKeys: String, CodingKey {
        case id, username, email
        case avatarUrl = "avatar_url"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct LoginRequest: Codable {
    let username: String
    let password: String
}

struct RegisterRequest: Codable {
    let username: String
    let email: String
    let password: String
    let confirmPassword: String
    
    enum CodingKeys: String, CodingKey {
        case username, email, password
        case confirmPassword = "confirm_password"
    }
}

struct AuthResponse: Codable {
    let success: Bool
    let message: String?
    let accessToken: String?
    let refreshToken: String?
    let user: User?
    let expiresIn: Int?
    
    enum CodingKeys: String, CodingKey {
        case success, message, user
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresIn = "expires_in"
    }
}

struct RefreshTokenRequest: Codable {
    let refreshToken: String
    
    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

// MARK: - 认证管理器

@MainActor
class AuthManager: ObservableObject {
    static let shared = AuthManager()
    
    @Published var isLoggedIn: Bool = false
    @Published var currentUser: User?
    @Published var isLoading: Bool = false
    
    private let httpClient: HttpClient
    private var authServerUrl: String = "http://localhost:8082"
    private var accessToken: String?
    private var refreshToken: String?
    private var autoRefreshEnabled: Bool = true
    private var refreshTimer: Timer?
    
    private let userDefaults = UserDefaults.standard
    private let keychainService = "com.taishanglaojun.desktop"
    
    private init() {
        self.httpClient = HttpClient.shared
        loadStoredAuthData()
    }
    
    deinit {
        refreshTimer?.invalidate()
    }
    
    // MARK: - 公共方法
    
    func setServerUrl(_ url: String) {
        authServerUrl = url
    }
    
    func enableAutoRefresh(_ enable: Bool) {
        autoRefreshEnabled = enable
        if enable && isLoggedIn {
            scheduleTokenRefresh()
        } else {
            refreshTimer?.invalidate()
        }
    }
    
    // MARK: - 认证方法
    
    func login(username: String, password: String) async throws -> AuthResponse {
        isLoading = true
        defer { isLoading = false }
        
        let request = LoginRequest(username: username, password: password)
        let url = "\(authServerUrl)/api/auth/login"
        
        let httpRequest = HttpRequest(
            url: url,
            method: .POST,
            headers: ["Content-Type": "application/json"],
            body: try JSONEncoder().encode(request)
        )
        
        let response = try await httpClient.send(request: httpRequest)
        
        guard response.statusCode == 200 else {
            throw AuthError.loginFailed("Login failed with status code: \(response.statusCode)")
        }
        
        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: response.data)
        
        if authResponse.success {
            await handleSuccessfulAuth(authResponse)
        }
        
        return authResponse
    }
    
    func register(username: String, email: String, password: String, confirmPassword: String) async throws -> AuthResponse {
        isLoading = true
        defer { isLoading = false }
        
        let request = RegisterRequest(
            username: username,
            email: email,
            password: password,
            confirmPassword: confirmPassword
        )
        let url = "\(authServerUrl)/api/auth/register"
        
        let httpRequest = HttpRequest(
            url: url,
            method: .POST,
            headers: ["Content-Type": "application/json"],
            body: try JSONEncoder().encode(request)
        )
        
        let response = try await httpClient.send(request: httpRequest)
        
        guard response.statusCode == 201 else {
            throw AuthError.registrationFailed("Registration failed with status code: \(response.statusCode)")
        }
        
        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: response.data)
        return authResponse
    }
    
    func logout() async throws {
        isLoading = true
        defer { isLoading = false }
        
        let url = "\(authServerUrl)/api/auth/logout"
        
        let httpRequest = HttpRequest(
            url: url,
            method: .POST,
            headers: getAuthHeaders()
        )
        
        let response = try await httpClient.send(request: httpRequest)
        
        // 无论服务器响应如何，都清除本地认证数据
        await clearAuthData()
        
        if response.statusCode != 200 {
            throw AuthError.logoutFailed("Logout failed with status code: \(response.statusCode)")
        }
    }
    
    func refreshAccessToken() async throws -> Bool {
        guard let refreshToken = refreshToken else {
            throw AuthError.noRefreshToken("No refresh token available")
        }
        
        let request = RefreshTokenRequest(refreshToken: refreshToken)
        let url = "\(authServerUrl)/api/auth/refresh"
        
        let httpRequest = HttpRequest(
            url: url,
            method: .POST,
            headers: ["Content-Type": "application/json"],
            body: try JSONEncoder().encode(request)
        )
        
        let response = try await httpClient.send(request: httpRequest)
        
        guard response.statusCode == 200 else {
            await clearAuthData()
            throw AuthError.tokenRefreshFailed("Token refresh failed with status code: \(response.statusCode)")
        }
        
        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: response.data)
        
        if authResponse.success, let newAccessToken = authResponse.accessToken {
            self.accessToken = newAccessToken
            updateAuthHeaders()
            saveAuthData()
            
            if autoRefreshEnabled {
                scheduleTokenRefresh()
            }
            
            return true
        }
        
        return false
    }
    
    // MARK: - Token管理
    
    func getAccessToken() -> String? {
        return accessToken
    }
    
    func getRefreshToken() -> String? {
        return refreshToken
    }
    
    private func getAuthHeaders() -> [String: String] {
        var headers: [String: String] = [:]
        if let token = accessToken {
            headers["Authorization"] = "Bearer \(token)"
        }
        return headers
    }
    
    private func updateAuthHeaders() {
        if let token = accessToken {
            httpClient.setDefaultHeader(key: "Authorization", value: "Bearer \(token)")
        }
    }
    
    // MARK: - 私有方法
    
    private func handleSuccessfulAuth(_ response: AuthResponse) async {
        self.accessToken = response.accessToken
        self.refreshToken = response.refreshToken
        self.currentUser = response.user
        self.isLoggedIn = true
        
        updateAuthHeaders()
        saveAuthData()
        
        if autoRefreshEnabled {
            scheduleTokenRefresh()
        }
    }
    
    private func clearAuthData() async {
        self.accessToken = nil
        self.refreshToken = nil
        self.currentUser = nil
        self.isLoggedIn = false
        
        httpClient.removeDefaultHeader(key: "Authorization")
        clearStoredAuthData()
        refreshTimer?.invalidate()
    }
    
    private func scheduleTokenRefresh() {
        refreshTimer?.invalidate()
        
        // 每50分钟刷新一次token（假设token有效期为1小时）
        refreshTimer = Timer.scheduledTimer(withTimeInterval: 50 * 60, repeats: true) { [weak self] _ in
            Task {
                try? await self?.refreshAccessToken()
            }
        }
    }
    
    // MARK: - 数据持久化
    
    private func saveAuthData() {
        if let user = currentUser {
            let userData = try? JSONEncoder().encode(user)
            userDefaults.set(userData, forKey: "current_user")
        }
        
        userDefaults.set(isLoggedIn, forKey: "is_logged_in")
        
        // 将敏感信息保存到Keychain
        if let accessToken = accessToken {
            saveToKeychain(key: "access_token", value: accessToken)
        }
        if let refreshToken = refreshToken {
            saveToKeychain(key: "refresh_token", value: refreshToken)
        }
    }
    
    private func loadStoredAuthData() {
        isLoggedIn = userDefaults.bool(forKey: "is_logged_in")
        
        if let userData = userDefaults.data(forKey: "current_user") {
            currentUser = try? JSONDecoder().decode(User.self, from: userData)
        }
        
        // 从Keychain加载敏感信息
        accessToken = loadFromKeychain(key: "access_token")
        refreshToken = loadFromKeychain(key: "refresh_token")
        
        if isLoggedIn && accessToken != nil {
            updateAuthHeaders()
            if autoRefreshEnabled {
                scheduleTokenRefresh()
            }
        }
    }
    
    private func clearStoredAuthData() {
        userDefaults.removeObject(forKey: "current_user")
        userDefaults.removeObject(forKey: "is_logged_in")
        
        deleteFromKeychain(key: "access_token")
        deleteFromKeychain(key: "refresh_token")
    }
    
    // MARK: - Keychain操作
    
    private func saveToKeychain(key: String, value: String) {
        let data = value.data(using: .utf8)!
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]
        
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }
    
    private func loadFromKeychain(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        guard status == errSecSuccess,
              let data = result as? Data,
              let string = String(data: data, encoding: .utf8) else {
            return nil
        }
        
        return string
    }
    
    private func deleteFromKeychain(key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key
        ]
        
        SecItemDelete(query as CFDictionary)
    }
}

// MARK: - 错误类型

enum AuthError: LocalizedError {
    case loginFailed(String)
    case registrationFailed(String)
    case logoutFailed(String)
    case tokenRefreshFailed(String)
    case noRefreshToken(String)
    case networkError(String)
    
    var errorDescription: String? {
        switch self {
        case .loginFailed(let message):
            return "Login failed: \(message)"
        case .registrationFailed(let message):
            return "Registration failed: \(message)"
        case .logoutFailed(let message):
            return "Logout failed: \(message)"
        case .tokenRefreshFailed(let message):
            return "Token refresh failed: \(message)"
        case .noRefreshToken(let message):
            return "No refresh token: \(message)"
        case .networkError(let message):
            return "Network error: \(message)"
        }
    }
}

// MARK: - 全局访问

extension AuthManager {
    static var isUserLoggedIn: Bool {
        return shared.isLoggedIn
    }
    
    static var currentUser: User? {
        return shared.currentUser
    }
    
    static func getAuthHeaders() -> [String: String] {
        return shared.getAuthHeaders()
    }
}