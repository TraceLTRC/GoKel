package gokel

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

const (
	RatingNotRated = iota
	RatingGeneral
	RatingTeen
	RatingMature
	RatingExplicit
)

const (
	ArchiveUnknownWarnings = (1 << iota)
	ArchiveNoWarnings
	ArchiveGraphicViolence
	ArchiveMajorDeath
	ArchiveNonCon
	ArchiveUnderage
)

const (
	CategoryGen = (1 << iota)
	CategoryFM
	CategoryMM
	CategoryMulti
	CategoryOther
	CategoryFF
)

const (
	PartPreface = iota
	PartStats
	PartChapters
)

const defaultWorkURL = "https://archiveofourown.org/works/%s?view_adult=true&view_full_work=true"

type WorkStats struct {
	PublishedTime   string
	LastUpdated     string
	Words           int
	CurrentChapters int
	MaxChapters     int
	Kudos           int
	Bookmarks       int
	Hits            int
}

type WorkMeta struct {
	WorkWarnings      int
	WorkRating        int
	WorkCategory      int
	WorkFandom        []string
	WorkRelationships []string
	WorkCharacters    []string
	WorkTags          []string
	WorkLanguange     string
	WorkStats         WorkStats
}

type WorkChapter struct {
	ChapterIndex          int
	ChapterTitle          string
	ChapterSummary        string
	ChapterBeginningNotes string
	ChapterEndingNotes    string
	ChapterContent        string
}

type WorkPreface struct {
	WorkTitle          string
	WorkAuthors        []string
	WorkSummary        string
	WorkBeginningNotes string
	WorkEndingNotes    string
	WorkSkin           string
}

type Work struct {
	WorkURL      string
	WorkChapters []WorkChapter
	WorkMeta
	WorkPreface
}

type Log struct {
	WorkId   string
	WorkPart int
	Payload  string
	Severity int
}

var collector *colly.Collector

func initCollector() {
	if collector != nil {
		return
	}

	collector = colly.NewCollector(
		colly.AllowedDomains("archiveofourown.org"),
	)

	collector.SetRequestTimeout(300 * time.Second)
}

func GetWorkURL(workID string) (workURL string) {
	return fmt.Sprintf(defaultWorkURL, workID)
}

func ParseChapterString(chapterString string) (chapters int, maxChapters int, err error) {
	splitStr := strings.SplitN(chapterString, "/", 2)

	chapters, err = strconv.Atoi(splitStr[0])
	if err != nil {
		return 0, 0, err
	}

	maxChapters, err = strconv.Atoi(splitStr[1])
	if err != nil {
		if splitStr[1] == "?" {
			maxChapters = -1
		} else {
			return 0, 0, err
		}
	}

	return chapters, maxChapters, nil
}

func GetRatingConstant(ratingString string) (rating int) {
	switch ratingString {
	case "Not Rated":
		return RatingNotRated
	case "General Audiences":
		return RatingGeneral
	case "Teen And Up Audiences":
		return RatingTeen
	case "Mature":
		return RatingMature
	case "Explicit":
		return RatingExplicit
	default:
		return RatingNotRated
	}
}

func GetWarningConstant(warningString string) (warning int) {
	switch warningString {
	case "Rape/Non-Con":
		return ArchiveNonCon
	case "Underage":
		return ArchiveUnderage
	case "Creator Chose Not To Use Archive Warnings":
		return ArchiveUnknownWarnings
	case "No Archive Warnings Apply":
		return ArchiveNoWarnings
	case "Graphic Depictions Of Violence":
		return ArchiveGraphicViolence
	default:
		return ArchiveUnknownWarnings
	}
}

func GetCategoryConstant(categoryString string) (category int) {
	switch categoryString {
	case "Gen":
		return CategoryGen
	case "F/M":
		return CategoryFM
	case "M/M":
		return CategoryFM
	case "Other":
		return CategoryOther
	case "F/F":
		return CategoryFF
	default:
		return 0
	}
}

func GetWork(workID string) (w Work, warns []Log, err error) {
	if collector == nil {
		initCollector()
	}

	w = Work{
		WorkURL: GetWorkURL(workID),
	}

	//Collect Prefaces
	collector.OnHTML("div#inner div#workskin > div.preface.group", func(h *colly.HTMLElement) {
		if h.DOM.HasClass("afterword") { // Is an afterword note
			workEndingNotes, err := h.DOM.ChildrenFiltered("div#work_endnotes").Children().Not("h3").Html()
			if err != nil {
				panic(fmt.Sprintf("Failed to get the end notes, %v", err))
			}

			w.WorkEndingNotes = strings.TrimSpace(workEndingNotes)
		} else {
			w.WorkTitle = strings.TrimSpace(h.DOM.ChildrenFiltered("h2").Text())

			h.DOM.ChildrenFiltered("h3").ChildrenFiltered("a").Each(func(_ int, s *goquery.Selection) {
				w.WorkAuthors = append(w.WorkAuthors, s.Text())
			})

			workSummary, err := h.DOM.ChildrenFiltered("div.summary").Children().Not("h3").Html()
			if err != nil {
				panic(fmt.Sprintf("Failed to get summary, %v", err))
			}
			w.WorkSummary = strings.TrimSpace(workSummary)

			workBeginningNotes, err := h.DOM.ChildrenFiltered("div.notes").Children().Not("h3").Not("p.jump").Html()
			if err != nil {
				panic(fmt.Sprintf("Failed to get the beginning notes, %v", err))
			}
			w.WorkBeginningNotes = strings.TrimSpace(workBeginningNotes)
		}
	})
	defer collector.OnHTMLDetach("div#inner div#workskin > div.preface.group")

	// Collect skinwork
	collector.OnHTML("div#inner style", func(h *colly.HTMLElement) {
		w.WorkSkin = strings.TrimSpace(h.Text)
	})
	defer collector.OnHTMLDetach("div#inner style")

	// Collect meta & stats
	collector.OnHTML("div#inner dl.work.meta", func(h *colly.HTMLElement) {
		h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
			classes := h.Attr("class")
			dataClass := strings.TrimSpace(strings.ReplaceAll(classes, "tags", ""))

			switch dataClass {
			case "rating":
				tag := h.DOM.Find("a").Text()
				w.WorkRating = GetRatingConstant(strings.TrimSpace(tag))
			case "warning":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkWarnings = w.WorkWarnings | GetWarningConstant(strings.TrimSpace(s.Text()))
				})
			case "category":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkCategory = w.WorkCategory | GetCategoryConstant(strings.TrimSpace(s.Text()))
				})
			case "fandom":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkFandom = append(w.WorkFandom, strings.TrimSpace(s.Text()))
				})
			case "relationship":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkRelationships = append(w.WorkRelationships, strings.TrimSpace(s.Text()))
				})
			case "character":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkCharacters = append(w.WorkCharacters, strings.TrimSpace(s.Text()))
				})
			case "freeform":
				h.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
					w.WorkTags = append(w.WorkTags, strings.TrimSpace(s.Text()))
				})
			case "language":
				w.WorkLanguange = strings.TrimSpace(h.Text)
			case "published":
				w.WorkStats.PublishedTime = strings.TrimSpace(h.Text)
			case "status":
				w.WorkStats.LastUpdated = strings.TrimSpace(h.Text)
			case "words":
				ret, err := strconv.Atoi(h.Text)
				if err != nil {
					panic(fmt.Sprintf("Failed to get word stats, %v", err))
				}
				w.WorkStats.Words = ret
			case "chapters":
				chapters, maxChapters, err := ParseChapterString(strings.TrimSpace(h.Text))
				if err != nil {
					panic(fmt.Sprintf("Failed to get chapter stats, %v", err))
				}
				w.WorkStats.CurrentChapters = chapters
				w.WorkStats.MaxChapters = maxChapters
			case "kudos":
				kudos, err := strconv.Atoi(strings.TrimSpace(h.Text))
				if err != nil {
					panic(fmt.Sprintf("Failed to get kudos stats, %v", err))
				}
				w.WorkStats.Kudos = kudos
			case "bookmarks":
				bookmarks, err := strconv.Atoi(strings.TrimSpace(h.Text))
				if err != nil {
					panic(fmt.Sprintf("Failed to get bookmark stats, %v", err))
				}
				w.WorkStats.Bookmarks = bookmarks
			case "hits":
				hits, err := strconv.Atoi(strings.TrimSpace(h.Text))
				if err != nil {
					panic(fmt.Sprintf("Failed to get hit stats, %v", err))
				}
				w.WorkStats.Hits = hits
			case "series", "stats", "comments":
				// Do nothing
			default:
				warns = append(warns, Log{
					WorkId:   workID,
					WorkPart: PartPreface,
					Payload:  fmt.Sprintf("Found an unhandled dataClass with name %s", dataClass),
				})
			}
		})
	})
	defer collector.OnHTMLDetach("div#inner dl.work.meta")

	// Collect chapters
	collector.OnHTML("div#inner div#chapters", func(h *colly.HTMLElement) {
		h.DOM.Children().Each(func(_ int, s *goquery.Selection) {
			elId := s.AttrOr("id", "")
			if !strings.HasPrefix(elId, "chapter") {
				return
			}
			chapId, err := strconv.Atoi(strings.TrimSpace(elId[8:]))
			if err != nil {
				panic(fmt.Sprintf("Unable to convert chapter index to number, %v", err))
			}

			workChapter := WorkChapter{
				ChapterIndex: chapId,
			}

			s.Children().Each(func(_ int, s *goquery.Selection) {
				if s.Is(".chapter.preface.group[role=\"complementary\"]") { // Beginning notes div & title
					chapTitle := strings.TrimSpace(s.ChildrenFiltered(".title").First().Clone().Children().Remove().End().Text())
					if strings.HasPrefix(chapTitle, ": ") {
						chapTitle = chapTitle[2:]
					} else if len(chapTitle) > 0 {
						warns = append(warns, Log{
							WorkId:   workID,
							WorkPart: PartPreface,
							Payload:  fmt.Sprintf("Chapter title contains unexpected prefix, %s", chapTitle),
						})
					}

					chapNotes, err := s.Find("div.summary > blockquote.userstuff").Html()
					if err != nil {
						panic(fmt.Sprintf("Unable to get chapter-%d notes, %v", chapId, err))
					}
					chapNotes = strings.TrimSpace(chapNotes)

					chapSummary, err := s.Find("div.notes > blockquote.userstuff").Html()
					if err != nil {
						panic(fmt.Sprintf("Unable to get chapter-%d summary, %v", chapId, err))
					}
					chapSummary = strings.TrimSpace(chapSummary)

					workChapter.ChapterTitle = chapTitle
					workChapter.ChapterBeginningNotes = chapNotes
					workChapter.ChapterSummary = chapSummary
				} else if s.Is(".userstuff.module") { // Content
					chapterContent, err := s.ChildrenFiltered(".landmark").Remove().End().Html()
					if err != nil {
						panic(fmt.Sprintf("Unable to get chapter-%d content, %v", chapId, err))
					}
					chapterContent = strings.TrimSpace(chapterContent)

					workChapter.ChapterContent = chapterContent
				} else if s.Is(".chapter.preface.group:not([role])") { // Ending notes
					chapterEndNotes, err := s.Find("blockquote").Html()
					if err != nil {
						panic(fmt.Sprintf("Unable to get chapter-%d end notes, %v", chapId, err))
					}
					chapterEndNotes = strings.TrimSpace(chapterEndNotes)

					workChapter.ChapterEndingNotes = chapterEndNotes
				} else {
					warns = append(warns, Log{
						WorkId:   workID,
						WorkPart: PartChapters,
						Payload:  fmt.Sprintf("Found an unhandleble element in chapter div, class: %s", s.AttrOr("class", "")),
					})
				}
			})

			w.WorkChapters = append(w.WorkChapters, workChapter)
		})
	})
	defer collector.OnHTMLDetach("div#inner div#chapters")

	err = collector.Visit(w.WorkURL)
	if err != nil {
		return w, warns, err
	}

	return w, warns, nil
}
