import Foundation
import Speech
import AVFoundation
import WatchKit
import Combine

/**
 * 语音报告管理器
 * 负责处理语音识别、语音合成和语音报告功能
 */
@MainActor
class VoiceReportManager: NSObject, ObservableObject {
    
    // MARK: - Published Properties
    
    @Published var isRecording = false
    @Published var isProcessing = false
    @Published var recognizedText = ""
    @Published var lastError: VoiceError?
    @Published var isAvailable = false
    @Published var recordingLevel: Float = 0.0
    
    // MARK: - Private Properties
    
    private let speechRecognizer: SFSpeechRecognizer?
    private let audioEngine = AVAudioEngine()
    private let speechSynthesizer = AVSpeechSynthesizer()
    
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    private var audioSession: AVAudioSession?
    
    private var cancellables = Set<AnyCancellable>()
    private var levelTimer: Timer?
    
    // MARK: - Configuration
    
    private let maxRecordingDuration: TimeInterval = 30.0
    private let silenceThreshold: Float = -40.0
    private let silenceDuration: TimeInterval = 2.0
    
    // MARK: - Initialization
    
    override init() {
        // 初始化语音识别器（中文）
        self.speechRecognizer = SFSpeechRecognizer(locale: Locale(identifier: "zh-CN"))
        
        super.init()
        
        setupAudioSession()
        checkAvailability()
        setupSpeechSynthesizer()
    }
    
    // MARK: - Public Methods
    
    /**
     * 开始语音录制
     */
    func startRecording() async throws {
        guard isAvailable else {
            throw VoiceError.notAvailable
        }
        
        guard !isRecording else {
            return
        }
        
        try await requestPermissions()
        
        // 停止之前的任务
        stopRecording()
        
        // 配置音频会话
        try configureAudioSession()
        
        // 创建识别请求
        recognitionRequest = SFSpeechAudioBufferRecognitionRequest()
        guard let recognitionRequest = recognitionRequest else {
            throw VoiceError.recognitionFailed("无法创建识别请求")
        }
        
        recognitionRequest.shouldReportPartialResults = true
        recognitionRequest.requiresOnDeviceRecognition = true
        
        // 配置音频引擎
        let inputNode = audioEngine.inputNode
        let recordingFormat = inputNode.outputFormat(forBus: 0)
        
        inputNode.installTap(onBus: 0, bufferSize: 1024, format: recordingFormat) { [weak self] buffer, _ in
            self?.recognitionRequest?.append(buffer)
            
            // 计算音频级别
            DispatchQueue.main.async {
                self?.updateRecordingLevel(buffer: buffer)
            }
        }
        
        // 启动音频引擎
        audioEngine.prepare()
        try audioEngine.start()
        
        // 开始识别
        recognitionTask = speechRecognizer?.recognitionTask(with: recognitionRequest) { [weak self] result, error in
            DispatchQueue.main.async {
                self?.handleRecognitionResult(result: result, error: error)
            }
        }
        
        isRecording = true
        recognizedText = ""
        lastError = nil
        
        // 启动录音级别监控
        startLevelMonitoring()
        
        // 设置最大录音时长
        DispatchQueue.main.asyncAfter(deadline: .now() + maxRecordingDuration) { [weak self] in
            if self?.isRecording == true {
                self?.stopRecording()
            }
        }
        
        // 触觉反馈
        WKInterfaceDevice.current().play(.start)
    }
    
    /**
     * 停止语音录制
     */
    func stopRecording() {
        guard isRecording else { return }
        
        // 停止音频引擎
        audioEngine.stop()
        audioEngine.inputNode.removeTap(onBus: 0)
        
        // 结束识别请求
        recognitionRequest?.endAudio()
        recognitionRequest = nil
        
        // 取消识别任务
        recognitionTask?.cancel()
        recognitionTask = nil
        
        isRecording = false
        recordingLevel = 0.0
        
        // 停止级别监控
        stopLevelMonitoring()
        
        // 触觉反馈
        WKInterfaceDevice.current().play(.stop)
    }
    
    /**
     * 语音播报文本
     */
    func speak(text: String, language: String = "zh-CN") async throws {
        guard !text.isEmpty else { return }
        
        return try await withCheckedThrowingContinuation { continuation in
            let utterance = AVSpeechUtterance(string: text)
            utterance.voice = AVSpeechSynthesisVoice(language: language)
            utterance.rate = AVSpeechUtteranceDefaultSpeechRate * 0.8
            utterance.volume = 0.8
            
            // 设置完成回调
            let delegate = SpeechDelegate { error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume()
                }
            }
            
            speechSynthesizer.delegate = delegate
            speechSynthesizer.speak(utterance)
        }
    }
    
    /**
     * 处理语音报告
     */
    func processVoiceReport() async throws -> VoiceReport {
        guard !recognizedText.isEmpty else {
            throw VoiceError.noTextRecognized
        }
        
        isProcessing = true
        defer { isProcessing = false }
        
        // 解析语音内容
        let report = try parseVoiceContent(recognizedText)
        
        // 清空识别文本
        recognizedText = ""
        
        return report
    }
    
    /**
     * 快速语音命令识别
     */
    func recognizeQuickCommand() async throws -> QuickCommand? {
        guard !recognizedText.isEmpty else { return nil }
        
        let text = recognizedText.lowercased()
        
        // 匹配快速命令
        for command in QuickCommand.allCases {
            if command.keywords.contains(where: { text.contains($0) }) {
                return command
            }
        }
        
        return nil
    }
    
    // MARK: - Private Methods
    
    private func setupAudioSession() {
        audioSession = AVAudioSession.sharedInstance()
        
        do {
            try audioSession?.setCategory(.playAndRecord, mode: .measurement, options: [.duckOthers])
            try audioSession?.setActive(true, options: .notifyOthersOnDeactivation)
        } catch {
            print("音频会话设置失败: \(error)")
        }
    }
    
    private func configureAudioSession() throws {
        guard let audioSession = audioSession else {
            throw VoiceError.audioSessionFailed
        }
        
        try audioSession.setCategory(.record, mode: .measurement, options: [.duckOthers])
        try audioSession.setActive(true, options: .notifyOthersOnDeactivation)
    }
    
    private func setupSpeechSynthesizer() {
        speechSynthesizer.delegate = self
    }
    
    private func checkAvailability() {
        guard let speechRecognizer = speechRecognizer else {
            isAvailable = false
            return
        }
        
        isAvailable = speechRecognizer.isAvailable
        
        // 监听可用性变化
        speechRecognizer.delegate = self
    }
    
    private func requestPermissions() async throws {
        // 请求语音识别权限
        let speechStatus = await SFSpeechRecognizer.requestAuthorization()
        guard speechStatus == .authorized else {
            throw VoiceError.permissionDenied("语音识别权限被拒绝")
        }
        
        // 请求麦克风权限
        let audioStatus = await AVAudioSession.sharedInstance().requestRecordPermission()
        guard audioStatus else {
            throw VoiceError.permissionDenied("麦克风权限被拒绝")
        }
    }
    
    private func handleRecognitionResult(result: SFSpeechRecognitionResult?, error: Error?) {
        if let error = error {
            lastError = VoiceError.recognitionFailed(error.localizedDescription)
            stopRecording()
            return
        }
        
        if let result = result {
            recognizedText = result.bestTranscription.formattedString
            
            // 如果识别完成，自动停止录音
            if result.isFinal {
                stopRecording()
            }
        }
    }
    
    private func updateRecordingLevel(buffer: AVAudioPCMBuffer) {
        guard let channelData = buffer.floatChannelData?[0] else { return }
        
        let channelDataArray = Array(UnsafeBufferPointer(start: channelData, count: Int(buffer.frameLength)))
        
        // 计算RMS值
        let rms = sqrt(channelDataArray.map { $0 * $0 }.reduce(0, +) / Float(channelDataArray.count))
        
        // 转换为分贝
        let db = 20 * log10(rms)
        
        // 标准化到0-1范围
        recordingLevel = max(0, min(1, (db + 60) / 60))
    }
    
    private func startLevelMonitoring() {
        levelTimer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { [weak self] _ in
            // 级别监控逻辑已在updateRecordingLevel中处理
        }
    }
    
    private func stopLevelMonitoring() {
        levelTimer?.invalidate()
        levelTimer = nil
    }
    
    private func parseVoiceContent(_ text: String) throws -> VoiceReport {
        // 简单的语音内容解析
        // 实际应用中可能需要更复杂的NLP处理
        
        let content = text.trimmingCharacters(in: .whitespacesAndNewlines)
        
        // 检测报告类型
        let reportType: VoiceReportType
        if content.contains("完成") || content.contains("结束") {
            reportType = .completion
        } else if content.contains("进度") || content.contains("更新") {
            reportType = .progress
        } else if content.contains("问题") || content.contains("困难") {
            reportType = .issue
        } else {
            reportType = .general
        }
        
        // 提取进度信息
        var progress: Double?
        let progressPattern = #"(\d+)%"#
        if let match = content.range(of: progressPattern, options: .regularExpression) {
            let progressString = String(content[match]).replacingOccurrences(of: "%", with: "")
            progress = Double(progressString).map { $0 / 100.0 }
        }
        
        return VoiceReport(
            id: UUID().uuidString,
            content: content,
            type: reportType,
            progress: progress,
            timestamp: Date(),
            confidence: 0.8 // 简化的置信度
        )
    }
}

// MARK: - SFSpeechRecognizerDelegate

extension VoiceReportManager: SFSpeechRecognizerDelegate {
    func speechRecognizer(_ speechRecognizer: SFSpeechRecognizer, availabilityDidChange available: Bool) {
        isAvailable = available
    }
}

// MARK: - AVSpeechSynthesizerDelegate

extension VoiceReportManager: AVSpeechSynthesizerDelegate {
    func speechSynthesizer(_ synthesizer: AVSpeechSynthesizer, didFinish utterance: AVSpeechUtterance) {
        // 语音播报完成
    }
    
    func speechSynthesizer(_ synthesizer: AVSpeechSynthesizer, didCancel utterance: AVSpeechUtterance) {
        // 语音播报取消
    }
}

// MARK: - Supporting Types

/**
 * 语音错误类型
 */
enum VoiceError: LocalizedError {
    case notAvailable
    case permissionDenied(String)
    case recognitionFailed(String)
    case audioSessionFailed
    case noTextRecognized
    
    var errorDescription: String? {
        switch self {
        case .notAvailable:
            return "语音识别不可用"
        case .permissionDenied(let message):
            return "权限被拒绝: \(message)"
        case .recognitionFailed(let message):
            return "识别失败: \(message)"
        case .audioSessionFailed:
            return "音频会话配置失败"
        case .noTextRecognized:
            return "未识别到文本"
        }
    }
}

/**
 * 语音报告数据模型
 */
struct VoiceReport: Identifiable, Codable {
    let id: String
    let content: String
    let type: VoiceReportType
    let progress: Double?
    let timestamp: Date
    let confidence: Double
}

/**
 * 语音报告类型
 */
enum VoiceReportType: String, CaseIterable, Codable {
    case completion = "completion"
    case progress = "progress"
    case issue = "issue"
    case general = "general"
    
    var displayName: String {
        switch self {
        case .completion: return "完成报告"
        case .progress: return "进度更新"
        case .issue: return "问题反馈"
        case .general: return "一般报告"
        }
    }
}

/**
 * 快速语音命令
 */
enum QuickCommand: String, CaseIterable {
    case acceptTask = "accept_task"
    case completeTask = "complete_task"
    case pauseTask = "pause_task"
    case syncData = "sync_data"
    case showTasks = "show_tasks"
    case showStats = "show_stats"
    
    var keywords: [String] {
        switch self {
        case .acceptTask:
            return ["接受", "开始", "领取"]
        case .completeTask:
            return ["完成", "结束", "提交"]
        case .pauseTask:
            return ["暂停", "停止", "中断"]
        case .syncData:
            return ["同步", "刷新", "更新"]
        case .showTasks:
            return ["显示任务", "查看任务", "任务列表"]
        case .showStats:
            return ["统计", "数据", "报告"]
        }
    }
    
    var displayName: String {
        switch self {
        case .acceptTask: return "接受任务"
        case .completeTask: return "完成任务"
        case .pauseTask: return "暂停任务"
        case .syncData: return "同步数据"
        case .showTasks: return "显示任务"
        case .showStats: return "显示统计"
        }
    }
}

/**
 * 语音合成代理
 */
private class SpeechDelegate: NSObject, AVSpeechSynthesizerDelegate {
    private let completion: (Error?) -> Void
    
    init(completion: @escaping (Error?) -> Void) {
        self.completion = completion
    }
    
    func speechSynthesizer(_ synthesizer: AVSpeechSynthesizer, didFinish utterance: AVSpeechUtterance) {
        completion(nil)
    }
    
    func speechSynthesizer(_ synthesizer: AVSpeechSynthesizer, didCancel utterance: AVSpeechUtterance) {
        completion(VoiceError.recognitionFailed("语音播报被取消"))
    }
}