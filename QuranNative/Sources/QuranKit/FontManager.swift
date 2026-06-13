import Foundation
import CoreGraphics
import CoreText

public final class FontManager: @unchecked Sendable {
    public static let shared = FontManager()

    public private(set) var amiriQuran: CTFont?
    public private(set) var amiriQuranColored: CTFont?
    public private(set) var amiriRegular: CTFont?
    public private(set) var noorehuda: CTFont?

    public var isAmiriLoaded: Bool { amiriQuran != nil }
    private let _lock = NSLock()
    private var _registered = false
    private var _loadErrors: [String] = []
    public var loadErrors: [String] { _lock.withLock { _loadErrors } }

    private init() {}

    public func registerAll(bundle: Bundle? = nil) {
        let b = bundle ?? Bundle.main
        _lock.withLock {
            guard !_registered else { return }
            _registered = true
            amiriQuran = registerFont(named: "AmiriQuran", size: 22, in: b)
            amiriQuranColored = registerFont(named: "AmiriQuranColored", size: 22, in: b)
            amiriRegular = registerFont(named: "Amiri-Regular", size: 22, in: b)
            noorehuda = registerFont(named: "noorehuda", size: 22, in: b)
        }
    }

    @discardableResult
    private func registerFont(named name: String, size: CGFloat, in bundle: Bundle) -> CTFont? {
        guard let url = locateFontURL(named: name, in: bundle) else {
            let msg = "font '\(name)' not found in bundle"
            _lock.withLock { _loadErrors.append(msg) }
            print("FontManager: \(msg)")
            return nil
        }
        var error: Unmanaged<CFError>?
        let success = CTFontManagerRegisterFontsForURL(url as CFURL, .process, &error)
        if !success {
            let err = error?.takeRetainedValue()
            let msg = "failed to register '\(name)': \(err?.localizedDescription ?? "unknown")"
            _lock.withLock { _loadErrors.append(msg) }
            print("FontManager: \(msg)")
            return nil
        }
        guard let descriptors = CTFontManagerCreateFontDescriptorsFromURL(url as CFURL) as? [CTFontDescriptor],
              let descriptor = descriptors.first else {
            let msg = "no font descriptors for '\(name)'"
            _lock.withLock { _loadErrors.append(msg) }
            print("FontManager: \(msg)")
            return nil
        }
        let font = CTFontCreateWithFontDescriptor(descriptor, size, nil)
        let family = CTFontCopyFamilyName(font) as String
        print("FontManager: loaded '\(name)' \u{2192} family='\(family)'")
        return font
    }

    private func locateFontURL(named name: String, in bundle: Bundle) -> URL? {
        if let url = bundle.url(forResource: name, withExtension: "ttf", subdirectory: "fonts") {
            return url
        }
        if let url = bundle.url(forResource: name, withExtension: "ttf") {
            return url
        }
        if let url = bundle.url(forResource: name, withExtension: "otf", subdirectory: "fonts") {
            return url
        }
        if let url = bundle.url(forResource: name, withExtension: "otf") {
            return url
        }
        return nil
    }

    /// Get a font for the given family name with a smart fallback chain:
    /// 1. Amiri Quran (the Mushaf standard, designed by Khaled Hosny)
    /// 2. Scheherazade New
    /// 3. SF Arabic (macOS system font)
    /// 4. Geeza Pro (macOS system font)
    public func quranFont(size: CGFloat) -> CTFont? {
        if let f = amiriQuran { return CTFontCreateCopyWithAttributes(f, size, nil, nil) }
        if let f = amiriQuranColored { return CTFontCreateCopyWithAttributes(f, size, nil, nil) }
        let families = [
            "Amiri Quran",
            "AmiriQuran",
            "Scheherazade New",
            "SF Arabic",
            "Geeza Pro",
            "Noto Naskh Arabic",
        ]
        for family in families {
            if let font = font(family: family, size: size) {
                return font
            }
        }
        return nil
    }

    public func font(family: String, size: CGFloat) -> CTFont? {
        let descriptor = CTFontDescriptorCreateWithAttributes([
            kCTFontFamilyNameAttribute as String: family,
        ] as CFDictionary)
        let font = CTFontCreateWithFontDescriptor(descriptor, size, nil)
        let fontName = CTFontCopyFamilyName(font) as String
        if fontName == family || fontName.hasPrefix(family) {
            return font
        }
        // Also try by PostScript name (some fonts are registered differently)
        let psDescriptor = CTFontDescriptorCreateWithAttributes([
            kCTFontNameAttribute as String: family,
        ] as CFDictionary)
        let psFont = CTFontCreateWithFontDescriptor(psDescriptor, size, nil)
        let psName = CTFontCopyFamilyName(psFont) as String
        if psName == family || psName.hasPrefix(family) || psName.contains("Amiri") {
            return psFont
        }
        return nil
    }
}
