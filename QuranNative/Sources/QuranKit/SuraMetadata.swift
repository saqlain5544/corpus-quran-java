import Foundation

public enum RevelationPlace: String, CaseIterable, Sendable {
    case makki, madani
}

private let REVELATION_PLACES: [Int: RevelationPlace] = {
    let makkiSurahs = Set([1,6,7,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,34,35,36,37,38,39,40,41,42,43,44,45,46,50,51,52,53,54,55,56,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,100,101,102,103,104,105,106,107,108,109,111,112,113,114])
    var result: [Int: RevelationPlace] = [:]
    for s in 1...114 { result[s] = makkiSurahs.contains(s) ? .makki : .madani }
    return result
}()

public var makkiSuras: [Int] {
    REVELATION_PLACES.filter { $0.value == .makki }.keys.sorted()
}
public var madaniSuras: [Int] {
    REVELATION_PLACES.filter { $0.value == .madani }.keys.sorted()
}
