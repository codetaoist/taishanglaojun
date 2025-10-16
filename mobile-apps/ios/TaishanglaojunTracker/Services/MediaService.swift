//
//  MediaService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import UIKit
import AVFoundation
import Photos
import Combine

/// 媒体服务管理器
class MediaService: NSObject, ObservableObject {
    
    // MARK: - Singleton
    static let shared = MediaService()
    
    // MARK: - Published Properties
    @Published var isRecording = false
    @Published var recordingDuration: TimeInterval = 0
    @Published var recordingError: String?
    
    // MARK: - Private Properties
    private var audioRecorder: AVAudioRecorder?
    private var audioPlayer: AVAudioPlayer?
    private var recordingTimer: Timer?
    private let audioSession = AVAudioSession.sharedInstance()
    
    // MARK: - File Management
    private let documentsDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
    
    private override init() {
        super.init()
        setupAudioSession()
    }
    
    // MARK: - Audio Session Setup
    private func setupAudioSession() {
        do {
            try audioSession.setCategory(.playAndRecord, mode: .default, options: [.defaultToSpeaker])
            try audioSession.setActive(true)
        } catch {
            print("❌ 音频会话设置失败: \(error)")
        }
    }
    
    // MARK: - Image Processing
    
    /// 压缩图片
    func compressImage(_ image: UIImage, quality: CGFloat = 0.7) -> Data? {
        return image.jpegData(compressionQuality: quality)
    }
    
    /// 调整图片大小
    func resizeImage(_ image: UIImage, targetSize: CGSize) -> UIImage? {
        let size = image.size
        
        let widthRatio = targetSize.width / size.width
        let heightRatio = targetSize.height / size.height
        
        // 保持宽高比
        let newSize: CGSize
        if widthRatio > heightRatio {
            newSize = CGSize(width: size.width * heightRatio, height: size.height * heightRatio)
        } else {
            newSize = CGSize(width: size.width * widthRatio, height: size.height * widthRatio)
        }
        
        let rect = CGRect(origin: .zero, size: newSize)
        
        UIGraphicsBeginImageContextWithOptions(newSize, false, 1.0)
        image.draw(in: rect)
        let newImage = UIGraphicsGetImageFromCurrentImageContext()
        UIGraphicsEndImageContext()
        
        return newImage
    }
    
    /// 保存图片到本地
    func saveImageToLocal(_ imageData: Data) -> URL? {
        let fileName = "image_\(UUID().uuidString).jpg"
        let fileURL = documentsDirectory.appendingPathComponent(fileName)
        
        do {
            try imageData.write(to: fileURL)
            return fileURL
        } catch {
            print("❌ 保存图片失败: \(error)")
            return nil
        }
    }
    
    /// 从本地加载图片
    func loadImageFromLocal(_ url: URL) -> UIImage? {
        guard let data = try? Data(contentsOf: url) else {
            return nil
        }
        return UIImage(data: data)
    }
    
    // MARK: - Audio Recording
    
    /// 开始录音
    func startRecording() -> AnyPublisher<Bool, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(MediaServiceError.serviceUnavailable))
                return
            }
            
            // 检查录音权限
            self.audioSession.requestRecordPermission { granted in
                DispatchQueue.main.async {
                    if granted {
                        self.beginRecording()
                        promise(.success(true))
                    } else {
                        promise(.failure(MediaServiceError.permissionDenied))
                    }
                }
            }
        }
        .eraseToAnyPublisher()
    }
    
    private func beginRecording() {
        let fileName = "audio_\(UUID().uuidString).m4a"
        let audioURL = documentsDirectory.appendingPathComponent(fileName)
        
        let settings = [
            AVFormatIDKey: Int(kAudioFormatMPEG4AAC),
            AVSampleRateKey: 44100,
            AVNumberOfChannelsKey: 2,
            AVEncoderAudioQualityKey: AVAudioQuality.high.rawValue
        ]
        
        do {
            audioRecorder = try AVAudioRecorder(url: audioURL, settings: settings)
            audioRecorder?.delegate = self
            audioRecorder?.record()
            
            isRecording = true
            recordingDuration = 0
            
            // 开始计时
            recordingTimer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { _ in
                self.recordingDuration += 0.1
            }
            
        } catch {
            recordingError = "录音失败: \(error.localizedDescription)"
            print("❌ 录音失败: \(error)")
        }
    }
    
    /// 停止录音
    func stopRecording() -> URL? {
        audioRecorder?.stop()
        recordingTimer?.invalidate()
        recordingTimer = nil
        
        isRecording = false
        
        return audioRecorder?.url
    }
    
    /// 播放音频
    func playAudio(from url: URL) -> AnyPublisher<Bool, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(MediaServiceError.serviceUnavailable))
                return
            }
            
            do {
                self.audioPlayer = try AVAudioPlayer(contentsOf: url)
                self.audioPlayer?.delegate = self
                self.audioPlayer?.play()
                promise(.success(true))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 停止播放
    func stopPlaying() {
        audioPlayer?.stop()
        audioPlayer = nil
    }
    
    // MARK: - File Management
    
    /// 删除本地文件
    func deleteLocalFile(at url: URL) {
        do {
            try FileManager.default.removeItem(at: url)
        } catch {
            print("❌ 删除文件失败: \(error)")
        }
    }
    
    /// 获取文件大小
    func getFileSize(at url: URL) -> Int64 {
        do {
            let attributes = try FileManager.default.attributesOfItem(atPath: url.path)
            return attributes[.size] as? Int64 ?? 0
        } catch {
            return 0
        }
    }
    
    // MARK: - Photo Library
    
    /// 检查相册权限
    func checkPhotoLibraryPermission() -> AnyPublisher<Bool, Error> {
        return Future { promise in
            let status = PHPhotoLibrary.authorizationStatus()
            
            switch status {
            case .authorized, .limited:
                promise(.success(true))
            case .denied, .restricted:
                promise(.failure(MediaServiceError.permissionDenied))
            case .notDetermined:
                PHPhotoLibrary.requestAuthorization { newStatus in
                    DispatchQueue.main.async {
                        promise(.success(newStatus == .authorized || newStatus == .limited))
                    }
                }
            @unknown default:
                promise(.failure(MediaServiceError.unknownError))
            }
        }
        .eraseToAnyPublisher()
    }
}

// MARK: - AVAudioRecorderDelegate
extension MediaService: AVAudioRecorderDelegate {
    func audioRecorderDidFinishRecording(_ recorder: AVAudioRecorder, successfully flag: Bool) {
        if !flag {
            recordingError = "录音未成功完成"
        }
    }
    
    func audioRecorderEncodeErrorDidOccur(_ recorder: AVAudioRecorder, error: Error?) {
        if let error = error {
            recordingError = "录音编码错误: \(error.localizedDescription)"
        }
    }
}

// MARK: - AVAudioPlayerDelegate
extension MediaService: AVAudioPlayerDelegate {
    func audioPlayerDidFinishPlaying(_ player: AVAudioPlayer, successfully flag: Bool) {
        audioPlayer = nil
    }
    
    func audioPlayerDecodeErrorDidOccur(_ player: AVAudioPlayer, error: Error?) {
        if let error = error {
            print("❌ 音频播放解码错误: \(error)")
        }
        audioPlayer = nil
    }
}

// MARK: - Media Service Errors
enum MediaServiceError: LocalizedError {
    case serviceUnavailable
    case permissionDenied
    case recordingFailed
    case playbackFailed
    case fileNotFound
    case compressionFailed
    case unknownError
    
    var errorDescription: String? {
        switch self {
        case .serviceUnavailable:
            return "媒体服务不可用"
        case .permissionDenied:
            return "权限被拒绝"
        case .recordingFailed:
            return "录音失败"
        case .playbackFailed:
            return "播放失败"
        case .fileNotFound:
            return "文件未找到"
        case .compressionFailed:
            return "压缩失败"
        case .unknownError:
            return "未知错误"
        }
    }
}