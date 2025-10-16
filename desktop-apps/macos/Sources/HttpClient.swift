import Foundation
import Network

// MARK: - HTTP Response Structure
public struct HttpResponse {
    public let statusCode: Int
    public let body: Data?
    public let headers: [String: String]
    public let success: Bool
    public let errorMessage: String?
    
    public var bodyString: String? {
        guard let body = body else { return nil }
        return String(data: body, encoding: .utf8)
    }
}

// MARK: - HTTP Request Structure
public struct HttpRequest {
    public let method: String
    public let url: String
    public let body: Data?
    public let headers: [String: String]
    public let timeoutInterval: TimeInterval
    
    public init(method: String, url: String, body: Data? = nil, 
                headers: [String: String] = [:], timeoutInterval: TimeInterval = 30.0) {
        self.method = method
        self.url = url
        self.body = body
        self.headers = headers
        self.timeoutInterval = timeoutInterval
    }
}

// MARK: - HTTP Client Class
public class HttpClient {
    private let session: URLSession
    private var baseURL: String?
    private var defaultHeaders: [String: String] = [:]
    
    // 全局HTTP客户端实例
    public static let shared = HttpClient()
    
    public init(configuration: URLSessionConfiguration = .default) {
        self.session = URLSession(configuration: configuration)
    }
    
    deinit {
        session.invalidateAndCancel()
    }
    
    // MARK: - Configuration Methods
    public func setBaseURL(_ baseURL: String) {
        self.baseURL = baseURL
    }
    
    public func setDefaultHeader(key: String, value: String) {
        defaultHeaders[key] = value
    }
    
    public func removeDefaultHeader(key: String) {
        defaultHeaders.removeValue(forKey: key)
    }
    
    // MARK: - Synchronous Request Method
    public func request(_ httpRequest: HttpRequest) -> HttpResponse {
        let semaphore = DispatchSemaphore(value: 0)
        var result: HttpResponse!
        
        requestAsync(httpRequest) { response in
            result = response
            semaphore.signal()
        }
        
        semaphore.wait()
        return result
    }
    
    // MARK: - Asynchronous Request Method
    public func requestAsync(_ httpRequest: HttpRequest, completion: @escaping (HttpResponse) -> Void) {
        guard let url = buildURL(from: httpRequest.url) else {
            let response = HttpResponse(
                statusCode: 0,
                body: nil,
                headers: [:],
                success: false,
                errorMessage: "Invalid URL"
            )
            completion(response)
            return
        }
        
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = httpRequest.method
        urlRequest.timeoutInterval = httpRequest.timeoutInterval
        
        // 设置默认头部
        for (key, value) in defaultHeaders {
            urlRequest.setValue(value, forHTTPHeaderField: key)
        }
        
        // 设置请求头部
        for (key, value) in httpRequest.headers {
            urlRequest.setValue(value, forHTTPHeaderField: key)
        }
        
        // 设置请求体
        if let body = httpRequest.body {
            urlRequest.httpBody = body
        }
        
        let task = session.dataTask(with: urlRequest) { data, response, error in
            let httpResponse = self.processResponse(data: data, response: response, error: error)
            DispatchQueue.main.async {
                completion(httpResponse)
            }
        }
        
        task.resume()
    }
    
    // MARK: - Convenience Methods
    public func get(_ url: String, headers: [String: String] = [:]) -> HttpResponse {
        let request = HttpRequest(method: "GET", url: url, headers: headers)
        return self.request(request)
    }
    
    public func post(_ url: String, body: Data? = nil, headers: [String: String] = [:]) -> HttpResponse {
        var requestHeaders = headers
        if body != nil && requestHeaders["Content-Type"] == nil {
            requestHeaders["Content-Type"] = "application/json"
        }
        
        let request = HttpRequest(method: "POST", url: url, body: body, headers: requestHeaders)
        return self.request(request)
    }
    
    public func post(_ url: String, jsonString: String, headers: [String: String] = [:]) -> HttpResponse {
        let body = jsonString.data(using: .utf8)
        return post(url, body: body, headers: headers)
    }
    
    public func put(_ url: String, body: Data? = nil, headers: [String: String] = [:]) -> HttpResponse {
        var requestHeaders = headers
        if body != nil && requestHeaders["Content-Type"] == nil {
            requestHeaders["Content-Type"] = "application/json"
        }
        
        let request = HttpRequest(method: "PUT", url: url, body: body, headers: requestHeaders)
        return self.request(request)
    }
    
    public func put(_ url: String, jsonString: String, headers: [String: String] = [:]) -> HttpResponse {
        let body = jsonString.data(using: .utf8)
        return put(url, body: body, headers: headers)
    }
    
    public func delete(_ url: String, headers: [String: String] = [:]) -> HttpResponse {
        let request = HttpRequest(method: "DELETE", url: url, headers: headers)
        return self.request(request)
    }
    
    // MARK: - Async/Await Methods (iOS 13.0+, macOS 10.15+)
    @available(macOS 10.15, *)
    public func requestAsync(_ httpRequest: HttpRequest) async -> HttpResponse {
        return await withCheckedContinuation { continuation in
            requestAsync(httpRequest) { response in
                continuation.resume(returning: response)
            }
        }
    }
    
    @available(macOS 10.15, *)
    public func getAsync(_ url: String, headers: [String: String] = [:]) async -> HttpResponse {
        let request = HttpRequest(method: "GET", url: url, headers: headers)
        return await requestAsync(request)
    }
    
    @available(macOS 10.15, *)
    public func postAsync(_ url: String, body: Data? = nil, headers: [String: String] = [:]) async -> HttpResponse {
        var requestHeaders = headers
        if body != nil && requestHeaders["Content-Type"] == nil {
            requestHeaders["Content-Type"] = "application/json"
        }
        
        let request = HttpRequest(method: "POST", url: url, body: body, headers: requestHeaders)
        return await requestAsync(request)
    }
    
    @available(macOS 10.15, *)
    public func postAsync(_ url: String, jsonString: String, headers: [String: String] = [:]) async -> HttpResponse {
        let body = jsonString.data(using: .utf8)
        return await postAsync(url, body: body, headers: headers)
    }
    
    // MARK: - Private Helper Methods
    private func buildURL(from urlString: String) -> URL? {
        if urlString.hasPrefix("http://") || urlString.hasPrefix("https://") {
            return URL(string: urlString)
        }
        
        guard let baseURL = baseURL else {
            return URL(string: urlString)
        }
        
        let fullURLString: String
        if baseURL.hasSuffix("/") && urlString.hasPrefix("/") {
            fullURLString = baseURL + String(urlString.dropFirst())
        } else if !baseURL.hasSuffix("/") && !urlString.hasPrefix("/") {
            fullURLString = baseURL + "/" + urlString
        } else {
            fullURLString = baseURL + urlString
        }
        
        return URL(string: fullURLString)
    }
    
    private func processResponse(data: Data?, response: URLResponse?, error: Error?) -> HttpResponse {
        if let error = error {
            return HttpResponse(
                statusCode: 0,
                body: nil,
                headers: [:],
                success: false,
                errorMessage: error.localizedDescription
            )
        }
        
        guard let httpResponse = response as? HTTPURLResponse else {
            return HttpResponse(
                statusCode: 0,
                body: data,
                headers: [:],
                success: false,
                errorMessage: "Invalid response type"
            )
        }
        
        var headers: [String: String] = [:]
        for (key, value) in httpResponse.allHeaderFields {
            if let keyString = key as? String, let valueString = value as? String {
                headers[keyString] = valueString
            }
        }
        
        let success = (200...299).contains(httpResponse.statusCode)
        
        return HttpResponse(
            statusCode: httpResponse.statusCode,
            body: data,
            headers: headers,
            success: success,
            errorMessage: success ? nil : "HTTP Error \(httpResponse.statusCode)"
        )
    }
}

// MARK: - JSON Helper Extensions
extension HttpClient {
    public func postJSON<T: Codable>(_ url: String, object: T, headers: [String: String] = [:]) -> HttpResponse {
        do {
            let jsonData = try JSONEncoder().encode(object)
            return post(url, body: jsonData, headers: headers)
        } catch {
            return HttpResponse(
                statusCode: 0,
                body: nil,
                headers: [:],
                success: false,
                errorMessage: "JSON encoding error: \(error.localizedDescription)"
            )
        }
    }
    
    public func putJSON<T: Codable>(_ url: String, object: T, headers: [String: String] = [:]) -> HttpResponse {
        do {
            let jsonData = try JSONEncoder().encode(object)
            return put(url, body: jsonData, headers: headers)
        } catch {
            return HttpResponse(
                statusCode: 0,
                body: nil,
                headers: [:],
                success: false,
                errorMessage: "JSON encoding error: \(error.localizedDescription)"
            )
        }
    }
}

extension HttpResponse {
    public func decodeJSON<T: Codable>(as type: T.Type) -> T? {
        guard let body = body else { return nil }
        
        do {
            return try JSONDecoder().decode(type, from: body)
        } catch {
            print("JSON decoding error: \(error)")
            return nil
        }
    }
}

// MARK: - Global Functions for C-style Interface
public func initHttpClient() -> Bool {
    // HttpClient.shared 已经在类定义中初始化
    return true
}

public func cleanupHttpClient() {
    // URLSession 会在 deinit 中自动清理
}

// MARK: - Network Monitoring (macOS 10.14+)
@available(macOS 10.14, *)
public class NetworkMonitor {
    private let monitor = NWPathMonitor()
    private let queue = DispatchQueue(label: "NetworkMonitor")
    
    public var isConnected: Bool = false
    public var connectionType: NWInterface.InterfaceType?
    
    public init() {
        startMonitoring()
    }
    
    private func startMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            self?.isConnected = path.status == .satisfied
            self?.connectionType = path.availableInterfaces.first?.type
        }
        monitor.start(queue: queue)
    }
    
    deinit {
        monitor.cancel()
    }
}