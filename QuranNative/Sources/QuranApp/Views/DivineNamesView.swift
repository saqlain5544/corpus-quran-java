import SwiftUI
import QuranKit

struct DivineNamesView: View {
    let names: [DivineName] = DIVINE_NAMES

    var body: some View {
        ScrollView {
            LazyVGrid(columns: [GridItem(.adaptive(minimum: 200), spacing: 16)], spacing: 16) {
                ForEach(names) { name in
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Text("\(name.number)")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .frame(width: 24, alignment: .trailing)

                            Text(name.arabic)
                                .font(.title3)
                                .frame(maxWidth: .infinity, alignment: .trailing)
                        }

                        Text(name.transliteration)
                            .font(.subheadline)
                            .foregroundStyle(.secondary)

                        Text(name.meaning)
                            .font(.headline)
                    }
                    .padding()
                    .background(.quaternary, in: RoundedRectangle(cornerRadius: 8))
                }
            }
            .padding()
        }
        .navigationTitle("99 Names of Allah")
    }
}
