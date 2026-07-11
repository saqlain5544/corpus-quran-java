# Quran Reader App - GO based

## Redo /root/details/[root:str] page

- Two Column layout: 1st column 5 grids, second column 3 grids
- First column has detailed meaning from `meanings-roots-ai.jsonl`
- Second column has detailed concordance data from `concordance.jsonl`
- Every `[surah:ayah] is parsed and linked to `surah/surah_number#verse`
- The page aligns with over aeshtetics of the app
- The arabic text whereever on the page is given `rtl direction`.
- All the occurances also show the exact ayah text from the quran as well.
- First column:

  |-----------------------------------------------------------|
  | Root (Arabic/Buckwalter) - <Total Occurance Count> |
  |-----------------------------------------------------------|
  | Description |
  | All Data from the data/morphology/meanings-roots-ai.jsonl |
  |-----------------------------------------------------------|

- Second column is fixed postion but doesn't go out of flow and looks like it is a sibling of frist column. it has overflow vertical and text wrap.
- The second column has lemmas distribution across the quran.
  |---------------------------------|
  | Lemma - <Total Occurance Count> |
  | ------------------------------- |
  | - Verse 1 with full text |
  | Transltion of the verse |
  | - Verse 2 |
  | and so on for every lemma |
  |---------------------------------|

## Tooltip Design

     --------------------
    |   Word uthmani     |
    |   Root(Ar/Buck)    |
    |   Lemma (from masaq)|
    |--------------------|
    |  Gloss of the word |
    |--------------------|
    | Morphology Tags    |
    |--------------------|
    | Grammatical Tags   |
     --------------------

## Surah Page

     ---------------------------------------------
    | Surah Header                                |
    |---------------------------------------------
    |Column1   |  Column Two                      |
    |          |  (rtl) Surah Text Verse by verse |
    |          | Translation after verse if       |
    |          | toggled on                       |
    |---------------------------------------------|
    | (fixed footer for search result navigation) |
    |              1/10 - <       >               |
    ----------------------------------------------

    						Column1
    	|---------------------------------------------|
    	|                   Word                      |
    	|             surah:verse:word_id             |
    	|---------------------------------------------|
    	|          Root (arabic/buckwalter)           |
    	|         Lemma (from masaq can be buckwalter)|
    	|---------------------------------------------|
    	|  Masaq Tags all from Masaq db               |
    	|---------------------------------------------|
    	| Full meaning data from the meaning_roots_ai |
    	|---------------------------------------------|
    	| Link to root/details/{root} page            |
    	|---------------------------------------------|

    	don't show cocordance data here just alll the masaq, and root meanings data. don't use multiple columns, use stacked data.

    - Search result navigation: when searching the surah all the results are sent to the browser. the browser holds (current surah and verse numbers where the results appear). on clicking next previous it scrolls to the surah and highlights for 2.5s. animates the scroll.
    - the search is just a copy of global search but it is different in that it doesn't navigate to another page and it doesn't change url at all. how is it similar is that it searches similarly to global search but it searches just the current surah and navigates using bottom controls.

    - Column1: this is a three col-grid layout that renders detailed root definitions only when clicked on a word. if the word has root and lemmas, it renders the definitions and sorts the current lemma on top in everything like frequency listing, concordance, shades of meaning, etc.
    - it is fixed and overflow vertical with proper word wrap.
    - every arabic line, word is properly direction rtl rendered.
    - Styled beautifly according to the theme.

- Surah Header:
  |-----------------------------------------------------------|
  | <Previous> Surah Name Translation Selector <Next> |
  |-----------------------------------------------------------|
  | <search_input> <font_controller> <line_height_controller> |  
  |-----------------------------------------------------------|

  Now the modal is no longer needed as its job is done by the sidebar.

  - Verse Number: an svg with text inside centered horizontally, vertically. the svg scales when the number of letters increase from let's say 1 to 10 then 100. max can be 3 digits. the verse number is after the verse text like this:
  - إِذۡ قَالَ لَهُۥ رَبُّهُۥٓ أَسۡلِمۡۖ قَالَ أَسۡلَمۡتُ لِرَبِّ ٱلۡعَٰلَمِينَ ١٣١

## Stack

- Vanilla GO
- Javascript
- HTML
- CSS

## Constraints

- Lowest Possible Memory Footprint
- Best UI design - utility classes + specific token classes, consistent theming, a detailed design system (typography, color scheme, grid system of 12 columns, for layout just usage of grid, for dynamic components, usage of javascript, proper stacking of components using proper z-indices, proper and consistent line-height, word spacing, etc)
- The app UI must be editorial styled.
- Only vanilla GO, Js, Css are allowed.

## Phases

### Plan

- Plan every component, module, struct and other layers before implementing anything.
- Decompose problems into atomic implementatable components.
- Create data flows, UI charts with clear interactions that users would perform.

### Pre Implementation

- Create pre implementation components, data pipelines, and other code run that using terminal in whatever language you choose
- Verify every module, function, etc with clear edge case handling and happy path handling
- If every thing works as intended, start implementation

### Test Driven Implementation

- Before implementing anything write detailed tests for each function/class, etc
- If every case is handled as planned, start writing the code
- The code must be atomic, modular, dry

## What we are building?

A Quran reading app that utilizes fully annotated text from Tanzil.net in xml format, MASAQ.csv, roots

- Annotated: pause marks, sajdah marks, ayah marks, rub' marks
- MASAQ.csv: a fully annotated Quran morphology in csv file

### Pages

- Homepage: lists all the surahs from 1st to 114, search box that takes to /search?q=rHM&type=root, etc., link to roots, link to concordance of roots.
- roots: lists roots paginated. first request returns the roots in order of frequency. has filtering: lowest freq to highest, highest to lowest. filtering: surah vise, just any surahs roots, synonym roots (if possible)
- root/detailed/[root:str]: lists all the occurances of the root in whereever they appear in the quran in the order from first surah to last. has filters: surah vise, frequency (asc, desc). has detailed meanings from root_meaning.csv.
- surah/[surah_number:int]: renders full surah in a paragraph form with hafs.woff2 fonts, every word has a tooltip attached to it that opens on hove and that is aligned such that its center is aligned with the center of the word and can be on top or bottom of the word depending on where the space is available and that fetches masaq and root data from the server.
  every word has a modal attached to it that opens on clicking and fetches detailed masaq data and links to the root/detailed/[root:str] page. a word may or may not have root. if a word doesn't have a root, it doesn't get the modal. - word: a word is a space delimited arabic entity with tashkeel. a pause mark or any other mark like end of surah, sajdah, rub', etc are not words.
- surah/[surah_number:int]/#[verse_number:int]: same as surah/[surah_number:int] only difference is that when on this router the view scrolls to that specific verse.
- search?q=[term=[term:str]&type=[type:str(root,english_term,arabic_word)]]: lists all the results related to the terms and type.
  - english_term: laveshtien or any other fuzzy matching english word or phrase against masaq glosses.
  - arabic_word: fuzzy match arabic word can be with tashkeel or without and can have different variants for example can be a Indo-Pak type, or Uthmani type or anyother. in IndoPak alif wasla is writter differently than in Uthmani, and vice versa.
  - properly themed and styled. renders full verse text that matches the result and highlight the word, phrase matched (arabic &/or translation)

### Major Components

- Global Header: has theme toggler. has the search_box with three selectors alongside the input box: root, english_term,arabic_word. Goes to search?q=[xxx] page as detailed above. has link to roots page and the homepage. stikcy top 0 relative to viewport (document of padding).
  - Theme: dark and light themes only no flashy colors except red error and blue success, etc.
- Surah Page Header that also is same for surah/surah_num/verse_num: has a local search box for just the surah similar features as the site search. this search just scrolls to first ayah matching the result. if the result doesn't match it just makes the text input border red error state. Has font size controller (font_size base 16px, range [10px-40px]). line_height controller (range: [1 to 3 in steps of 0.1]) base line height 1.6. this is sticky below the global header. top 0 relative to global header. has next surah, previous surah if any (on first suran previous is disabled on last surah next is disabled). input sliders have proper theme matching styling, typography, etc.

## Translation Integration

- Add English translation(s) from ltr rendering
- Add Urdu translation(s) from rtl
- Must match the textual format of verses
- Transliterition is not added to the ui but kept at the backend for fuzzy searching.

## Search Engine

- Fuzzy match translations, glosses, arabic text (with or without tashkeel), transliterition and full match roots (buckwalter, arabic with spaces without spaces, with tashkeel without tashkeel), fuzzy match lemmas (arabic with or without tashkeel, transilterition)

## Important Notes

- Every feature is first researched using all the available tools in opencode, and internet.
- Everything is build in steps
- Everything is first build on paper in docs/pen-and-paper with detailed flowcharts, and other visaulizations
- Everything has full state of every function/component/part in a to do form in docs/to-dos and then keep updating those to-dos
- Every data flow is built on paper, tested with scripts using terminal
- Only than you can implement
- Use subagents for researching, testing, debugging, implementation.

## Data

./data folder has all the data

- Trusted data: ./data/fonts/hafs.woff2, ./data/morphology/MASAQ.csv, ./data/quran/quran-uthmani.xml. these are scholarly data.
- Other data needs verification and testing for accuracy

## What if not enough information provided or confused?

- Ask clarifying questions.
- Use internet and other sources for consultation and research
- Use context7 for documentation if that is not working use internet directly.

## What not to do?

Don't make assumptions, don't hallucinate, don't write untested code and don't utilize untrustworthy data without mentioning explicitly in the UI.
