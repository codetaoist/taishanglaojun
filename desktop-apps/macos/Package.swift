// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "TaishanglaojunDesktop",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "TaishanglaojunDesktop",
            targets: ["TaishanglaojunDesktop"]
        )
    ],
    dependencies: [
        // 网络请求
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
        // JSON处理
        .package(url: "https://github.com/SwiftyJSON/SwiftyJSON.git", from: "5.0.0"),
        // WebSocket
        .package(url: "https://github.com/daltoniam/Starscream.git", from: "4.0.0"),
        // 加密
        .package(url: "https://github.com/krzyzanowskim/CryptoSwift.git", from: "1.8.0"),
        // 日志
        .package(url: "https://github.com/apple/swift-log.git", from: "1.5.0"),
    ],
    targets: [
        .executableTarget(
            name: "TaishanglaojunDesktop",
            dependencies: [
                "Alamofire",
                "SwiftyJSON", 
                "Starscream",
                "CryptoSwift",
                .product(name: "Logging", package: "swift-log"),
            ],
            path: "Sources",
            resources: [
                .process("Resources")
            ]
        ),
        .testTarget(
            name: "TaishanglaojunDesktopTests",
            dependencies: ["TaishanglaojunDesktop"],
            path: "Tests"
        )
    ]
)