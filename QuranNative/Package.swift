// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "QuranNative",
    platforms: [.macOS(.v14)],
    products: [
        .library(name: "QuranKit", targets: ["QuranKit"]),
        .executable(name: "QuranApp", targets: ["QuranApp"]),
    ],
    targets: [
        .target(
            name: "QuranKit",
            dependencies: [],
            path: "Sources/QuranKit",
            resources: []
        ),
        .executableTarget(
            name: "QuranApp",
            dependencies: ["QuranKit"],
            path: "Sources/QuranApp",
            resources: [
                .copy("Resources")
            ]
        ),
        .testTarget(
            name: "QuranKitTests",
            dependencies: ["QuranKit"],
            path: "Tests/QuranKitTests",
            resources: []
        ),
    ]
)
