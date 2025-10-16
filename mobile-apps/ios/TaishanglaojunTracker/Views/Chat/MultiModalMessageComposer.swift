//
//  MultiModalMessageComposer.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import SwiftUI
import PhotosUI
import AVFoundation

/// 多模态消息编辑器
struct MultiModalMessageComposer: View {
    
    // MARK: - Properties
    @StateObject private var mediaService = MediaService.shared
    @State private var messageText = ""
    @State private var selectedImage: UIImage?
    @State private var selectedImageData: Data?
    @State private var audioURL: URL?
    @State private var showingImagePicker = false
    @State private var showingCamera = false
    @State private var showingActionSheet = false
    @State private var isRecording = false
    @State private var recordingDuration: TimeInterval = 0
    @State private var showingAlert = false
    @State private var alertMessage = ""
    
    // MARK: - Callbacks
    let onSendText: (String) -> Void
    let onSendImage: (Data) -> Void
    let onSendAudio: (URL) -> Void
    
    // MARK: - Body
    var body: some View {
        VStack(spacing: 12) {
            // 媒体预览区域
            if selectedImage != nil || audioURL != nil {
                mediaPreviewSection
            }
            
            // 消息输入区域
            messageInputSection
            
            // 操作按钮区域
            actionButtonsSection
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(16)
        .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
        .sheet(isPresented: $showingImagePicker) {
            ImagePicker(image: $selectedImage, imageData: $selectedImageData)
        }
        .sheet(isPresented: $showingCamera) {
            CameraPicker(image: $selectedImage, imageData: $selectedImageData)
        }
        .actionSheet(isPresented: $showingActionSheet) {
            ActionSheet(
                title: Text("选择图片"),
                buttons: [
                    .default(Text("相机")) {
                        showingCamera = true
                    },
                    .default(Text("相册")) {
                        showingImagePicker = true
                    },
                    .cancel()
                ]
            )
        }
        .alert("提示", isPresented: $showingAlert) {
            Button("确定") { }
        } message: {
            Text(alertMessage)
        }
        .onReceive(mediaService.$recordingDuration) { duration in
            recordingDuration = duration
        }
        .onReceive(mediaService.$isRecording) { recording in
            isRecording = recording
        }
        .onReceive(mediaService.$recordingError) { error in
            if let error = error {
                alertMessage = error
                showingAlert = true
            }
        }
    }
    
    // MARK: - Media Preview Section
    private var mediaPreviewSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("媒体预览")
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                Spacer()
                
                Button("清除") {
                    clearMedia()
                }
                .font(.caption)
                .foregroundColor(.red)
            }
            
            if let image = selectedImage {
                // 图片预览
                Image(uiImage: image)
                    .resizable()
                    .aspectRatio(contentMode: .fit)
                    .frame(maxHeight: 200)
                    .cornerRadius(8)
            }
            
            if let audioURL = audioURL {
                // 音频预览
                HStack {
                    Image(systemName: "waveform")
                        .foregroundColor(.blue)
                    
                    Text("音频消息")
                        .font(.body)
                    
                    Spacer()
                    
                    Button(action: {
                        playAudio(audioURL)
                    }) {
                        Image(systemName: "play.circle.fill")
                            .font(.title2)
                            .foregroundColor(.blue)
                    }
                }
                .padding()
                .background(Color(.systemGray6))
                .cornerRadius(8)
            }
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    // MARK: - Message Input Section
    private var messageInputSection: some View {
        HStack(alignment: .bottom, spacing: 12) {
            // 文本输入框
            TextField("输入消息...", text: $messageText, axis: .vertical)
                .textFieldStyle(RoundedBorderTextFieldStyle())
                .lineLimit(1...5)
            
            // 发送按钮
            Button(action: sendMessage) {
                Image(systemName: "paperplane.fill")
                    .font(.title2)
                    .foregroundColor(.white)
                    .frame(width: 36, height: 36)
                    .background(canSend ? Color.blue : Color.gray)
                    .clipShape(Circle())
            }
            .disabled(!canSend)
        }
    }
    
    // MARK: - Action Buttons Section
    private var actionButtonsSection: some View {
        HStack(spacing: 20) {
            // 图片按钮
            Button(action: {
                showingActionSheet = true
            }) {
                VStack(spacing: 4) {
                    Image(systemName: "photo")
                        .font(.title2)
                        .foregroundColor(.blue)
                    Text("图片")
                        .font(.caption)
                        .foregroundColor(.blue)
                }
            }
            
            Spacer()
            
            // 录音按钮
            Button(action: toggleRecording) {
                VStack(spacing: 4) {
                    ZStack {
                        Circle()
                            .fill(isRecording ? Color.red : Color.blue)
                            .frame(width: 40, height: 40)
                        
                        Image(systemName: isRecording ? "stop.fill" : "mic.fill")
                            .font(.title2)
                            .foregroundColor(.white)
                        
                        if isRecording {
                            Circle()
                                .stroke(Color.red, lineWidth: 2)
                                .frame(width: 50, height: 50)
                                .scaleEffect(isRecording ? 1.2 : 1.0)
                                .animation(.easeInOut(duration: 0.5).repeatForever(), value: isRecording)
                        }
                    }
                    
                    Text(isRecording ? String(format: "%.1fs", recordingDuration) : "录音")
                        .font(.caption)
                        .foregroundColor(isRecording ? .red : .blue)
                }
            }
            
            Spacer()
            
            // 更多选项按钮
            Button(action: {
                // 可以添加更多功能，如文件、位置等
            }) {
                VStack(spacing: 4) {
                    Image(systemName: "plus.circle")
                        .font(.title2)
                        .foregroundColor(.blue)
                    Text("更多")
                        .font(.caption)
                        .foregroundColor(.blue)
                }
            }
        }
        .padding(.horizontal)
    }
    
    // MARK: - Computed Properties
    private var canSend: Bool {
        !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ||
        selectedImageData != nil ||
        audioURL != nil
    }
    
    // MARK: - Methods
    
    /// 发送消息
    private func sendMessage() {
        // 发送文本消息
        if !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            onSendText(messageText)
            messageText = ""
        }
        
        // 发送图片消息
        if let imageData = selectedImageData {
            onSendImage(imageData)
            selectedImage = nil
            selectedImageData = nil
        }
        
        // 发送音频消息
        if let audioURL = audioURL {
            onSendAudio(audioURL)
            self.audioURL = nil
        }
    }
    
    /// 清除媒体
    private func clearMedia() {
        selectedImage = nil
        selectedImageData = nil
        audioURL = nil
    }
    
    /// 切换录音状态
    private func toggleRecording() {
        if isRecording {
            // 停止录音
            if let url = mediaService.stopRecording() {
                audioURL = url
            }
        } else {
            // 开始录音
            mediaService.startRecording()
                .sink(
                    receiveCompletion: { completion in
                        if case .failure(let error) = completion {
                            alertMessage = error.localizedDescription
                            showingAlert = true
                        }
                    },
                    receiveValue: { _ in }
                )
                .store(in: &cancellables)
        }
    }
    
    /// 播放音频
    private func playAudio(_ url: URL) {
        mediaService.playAudio(from: url)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        alertMessage = "播放失败: \(error.localizedDescription)"
                        showingAlert = true
                    }
                },
                receiveValue: { _ in }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Combine
    @State private var cancellables = Set<AnyCancellable>()
}

// MARK: - Image Picker
struct ImagePicker: UIViewControllerRepresentable {
    @Binding var image: UIImage?
    @Binding var imageData: Data?
    @Environment(\.presentationMode) var presentationMode
    
    func makeUIViewController(context: Context) -> PHPickerViewController {
        var config = PHPickerConfiguration()
        config.filter = .images
        config.selectionLimit = 1
        
        let picker = PHPickerViewController(configuration: config)
        picker.delegate = context.coordinator
        return picker
    }
    
    func updateUIViewController(_ uiViewController: PHPickerViewController, context: Context) {}
    
    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }
    
    class Coordinator: NSObject, PHPickerViewControllerDelegate {
        let parent: ImagePicker
        
        init(_ parent: ImagePicker) {
            self.parent = parent
        }
        
        func picker(_ picker: PHPickerViewController, didFinishPicking results: [PHPickerResult]) {
            parent.presentationMode.wrappedValue.dismiss()
            
            guard let provider = results.first?.itemProvider else { return }
            
            if provider.canLoadObject(ofClass: UIImage.self) {
                provider.loadObject(ofClass: UIImage.self) { image, _ in
                    DispatchQueue.main.async {
                        if let uiImage = image as? UIImage {
                            self.parent.image = uiImage
                            self.parent.imageData = MediaService.shared.compressImage(uiImage)
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Camera Picker
struct CameraPicker: UIViewControllerRepresentable {
    @Binding var image: UIImage?
    @Binding var imageData: Data?
    @Environment(\.presentationMode) var presentationMode
    
    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.delegate = context.coordinator
        picker.sourceType = .camera
        picker.allowsEditing = true
        return picker
    }
    
    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}
    
    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }
    
    class Coordinator: NSObject, UINavigationControllerDelegate, UIImagePickerControllerDelegate {
        let parent: CameraPicker
        
        init(_ parent: CameraPicker) {
            self.parent = parent
        }
        
        func imagePickerController(_ picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey : Any]) {
            if let uiImage = info[.editedImage] as? UIImage ?? info[.originalImage] as? UIImage {
                parent.image = uiImage
                parent.imageData = MediaService.shared.compressImage(uiImage)
            }
            
            parent.presentationMode.wrappedValue.dismiss()
        }
        
        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            parent.presentationMode.wrappedValue.dismiss()
        }
    }
}

// MARK: - Preview
struct MultiModalMessageComposer_Previews: PreviewProvider {
    static var previews: some View {
        MultiModalMessageComposer(
            onSendText: { _ in },
            onSendImage: { _ in },
            onSendAudio: { _ in }
        )
        .padding()
    }
}