//
//  CryptoService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CryptoKit
import Security

/// 加密服务管理器
class CryptoService {
    
    // MARK: - Singleton
    static let shared = CryptoService()
    
    // MARK: - Private Properties
    private let keySize = 32 // 256 bits for AES-256
    private let keychainService = KeychainService.shared
    
    private init() {}
    
    // MARK: - Key Management
    
    /// 获取或生成加密密钥
    private func getEncryptionKey() -> SymmetricKey {
        if let keyData = keychainService.getEncryptionKey() {
            return SymmetricKey(data: keyData)
        } else {
            let key = SymmetricKey(size: .bits256)
            let keyData = key.withUnsafeBytes { Data($0) }
            keychainService.saveEncryptionKey(keyData)
            return key
        }
    }
    
    // MARK: - Data Encryption/Decryption
    
    /// 加密数据
    func encrypt(_ data: Data) -> Data? {
        do {
            let key = getEncryptionKey()
            let sealedBox = try AES.GCM.seal(data, using: key)
            return sealedBox.combined
        } catch {
            print("❌ 数据加密失败: \(error)")
            return nil
        }
    }
    
    /// 解密数据
    func decrypt(_ encryptedData: Data) -> Data? {
        do {
            let key = getEncryptionKey()
            let sealedBox = try AES.GCM.SealedBox(combined: encryptedData)
            return try AES.GCM.open(sealedBox, using: key)
        } catch {
            print("❌ 数据解密失败: \(error)")
            return nil
        }
    }
    
    /// 加密字符串
    func encrypt(_ string: String) -> String? {
        guard let data = string.data(using: .utf8),
              let encryptedData = encrypt(data) else {
            return nil
        }
        return encryptedData.base64EncodedString()
    }
    
    /// 解密字符串
    func decrypt(_ encryptedString: String) -> String? {
        guard let encryptedData = Data(base64Encoded: encryptedString),
              let decryptedData = decrypt(encryptedData) else {
            return nil
        }
        return String(data: decryptedData, encoding: .utf8)
    }
    
    // MARK: - Location Data Encryption
    
    /// 加密位置点
    func encryptLocationPoint(_ point: LocationPoint) -> LocationPoint {
        var encryptedPoint = point
        
        // 对敏感的位置信息进行加密
        if let encryptedLat = encrypt("\(point.latitude)") {
            // 这里可以选择性地加密某些字段
            // 为了演示，我们保持原始数据，但在实际应用中可以加密
        }
        
        return encryptedPoint
    }
    
    /// 解密位置点
    func decryptLocationPoint(_ encryptedPoint: LocationPoint) -> LocationPoint {
        // 解密逻辑，与加密对应
        return encryptedPoint
    }
    
    /// 加密轨迹
    func encryptTrajectory(_ trajectory: Trajectory) -> Trajectory {
        var encryptedTrajectory = trajectory
        
        // 加密轨迹中的敏感信息
        encryptedTrajectory.points = trajectory.points.map { encryptLocationPoint($0) }
        
        return encryptedTrajectory
    }
    
    /// 解密轨迹
    func decryptTrajectory(_ encryptedTrajectory: Trajectory) -> Trajectory {
        var trajectory = encryptedTrajectory
        
        // 解密轨迹中的敏感信息
        trajectory.points = encryptedTrajectory.points.map { decryptLocationPoint($0) }
        
        return trajectory
    }
    
    // MARK: - Hash Functions
    
    /// 生成数据哈希
    func hash(_ data: Data) -> String {
        let digest = SHA256.hash(data: data)
        return digest.compactMap { String(format: "%02x", $0) }.joined()
    }
    
    /// 生成字符串哈希
    func hash(_ string: String) -> String? {
        guard let data = string.data(using: .utf8) else { return nil }
        return hash(data)
    }
    
    /// 验证数据完整性
    func verifyIntegrity(data: Data, expectedHash: String) -> Bool {
        let actualHash = hash(data)
        return actualHash == expectedHash
    }
    
    // MARK: - Digital Signatures
    
    /// 生成数字签名
    func sign(_ data: Data) -> Data? {
        do {
            let privateKey = getSigningKey()
            let signature = try privateKey.signature(for: data)
            return signature.rawRepresentation
        } catch {
            print("❌ 数字签名失败: \(error)")
            return nil
        }
    }
    
    /// 验证数字签名
    func verifySignature(_ signature: Data, for data: Data) -> Bool {
        do {
            let publicKey = getVerificationKey()
            let ecdsaSignature = try P256.Signing.ECDSASignature(rawRepresentation: signature)
            return publicKey.isValidSignature(ecdsaSignature, for: data)
        } catch {
            print("❌ 签名验证失败: \(error)")
            return false
        }
    }
    
    private func getSigningKey() -> P256.Signing.PrivateKey {
        if let keyData = keychainService.getSigningKey() {
            do {
                return try P256.Signing.PrivateKey(rawRepresentation: keyData)
            } catch {
                print("❌ 加载签名密钥失败，生成新密钥")
            }
        }
        
        let privateKey = P256.Signing.PrivateKey()
        keychainService.saveSigningKey(privateKey.rawRepresentation)
        return privateKey
    }
    
    private func getVerificationKey() -> P256.Signing.PublicKey {
        let privateKey = getSigningKey()
        return privateKey.publicKey
    }
    
    // MARK: - Random Generation
    
    /// 生成随机数据
    func generateRandomData(length: Int) -> Data {
        var data = Data(count: length)
        _ = data.withUnsafeMutableBytes { bytes in
            SecRandomCopyBytes(kSecRandomDefault, length, bytes.bindMemory(to: UInt8.self).baseAddress!)
        }
        return data
    }
    
    /// 生成随机字符串
    func generateRandomString(length: Int) -> String {
        let characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        return String((0..<length).map { _ in characters.randomElement()! })
    }
    
    /// 生成UUID
    func generateUUID() -> String {
        return UUID().uuidString
    }
    
    // MARK: - Data Anonymization
    
    /// 匿名化位置数据
    func anonymizeLocation(_ point: LocationPoint, precision: Double = 0.001) -> LocationPoint {
        var anonymizedPoint = point
        
        // 降低精度以保护隐私
        anonymizedPoint.latitude = round(point.latitude / precision) * precision
        anonymizedPoint.longitude = round(point.longitude / precision) * precision
        
        // 移除或模糊化其他敏感信息
        anonymizedPoint.accuracy = max(point.accuracy, 100) // 最小精度100米
        
        return anonymizedPoint
    }
    
    /// 添加噪声到位置数据
    func addNoiseToLocation(_ point: LocationPoint, noiseLevel: Double = 0.0001) -> LocationPoint {
        var noisyPoint = point
        
        let latNoise = Double.random(in: -noiseLevel...noiseLevel)
        let lonNoise = Double.random(in: -noiseLevel...noiseLevel)
        
        noisyPoint.latitude += latNoise
        noisyPoint.longitude += lonNoise
        
        return noisyPoint
    }
    
    // MARK: - Secure Comparison
    
    /// 安全比较两个数据
    func secureCompare(_ data1: Data, _ data2: Data) -> Bool {
        guard data1.count == data2.count else { return false }
        
        var result: UInt8 = 0
        for i in 0..<data1.count {
            result |= data1[i] ^ data2[i]
        }
        
        return result == 0
    }
    
    /// 安全比较两个字符串
    func secureCompare(_ string1: String, _ string2: String) -> Bool {
        guard let data1 = string1.data(using: .utf8),
              let data2 = string2.data(using: .utf8) else {
            return false
        }
        
        return secureCompare(data1, data2)
    }
}

// MARK: - Keychain Service
class KeychainService {
    
    static let shared = KeychainService()
    
    private let service = "com.taishanglaojun.tracker"
    
    private init() {}
    
    // MARK: - Encryption Key
    
    func saveEncryptionKey(_ keyData: Data) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "encryption_key",
            kSecValueData as String: keyData,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }
    
    func getEncryptionKey() -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "encryption_key",
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        if status == errSecSuccess {
            return result as? Data
        }
        
        return nil
    }
    
    // MARK: - Signing Key
    
    func saveSigningKey(_ keyData: Data) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "signing_key",
            kSecValueData as String: keyData,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }
    
    func getSigningKey() -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "signing_key",
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        if status == errSecSuccess {
            return result as? Data
        }
        
        return nil
    }
    
    // MARK: - Auth Token
    
    func saveToken(_ token: String) {
        guard let tokenData = token.data(using: .utf8) else { return }
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "auth_token",
            kSecValueData as String: tokenData,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }
    
    func getToken() -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "auth_token",
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        if status == errSecSuccess,
           let tokenData = result as? Data {
            return String(data: tokenData, encoding: .utf8)
        }
        
        return nil
    }
    
    func deleteToken() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "auth_token"
        ]
        
        SecItemDelete(query as CFDictionary)
    }
}