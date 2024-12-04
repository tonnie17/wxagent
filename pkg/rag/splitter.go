package rag

import (
	"strings"
	"unicode/utf8"
)

func splitText(text string, chunkSize int, chunkOverlap int) []string {
	separators := []string{"\n\n", "\n", " ", ""}
	separator := separators[len(separators)-1]

	var newSeparators []string
	for i, sep := range separators {
		if sep == "" || strings.Contains(text, sep) {
			separator = sep
			newSeparators = separators[i+1:]
			break
		}
	}

	var final []string
	goodSplits := make([]string, 0)
	for _, split := range strings.Split(text, separator) {
		if utf8.RuneCountInString(split) < chunkSize {
			goodSplits = append(goodSplits, split)
			continue
		}

		if len(goodSplits) > 0 {
			mergedSplits := mergeSplits(goodSplits, separator, chunkSize, chunkOverlap)

			final = append(final, mergedSplits...)
			goodSplits = make([]string, 0)
		}

		if len(newSeparators) == 0 {
			final = append(final, split)
		} else {
			other := splitText(split, chunkSize, chunkOverlap)
			final = append(final, other...)
		}
	}

	if len(goodSplits) > 0 {
		mergedSplits := mergeSplits(goodSplits, separator, chunkSize, chunkOverlap)
		final = append(final, mergedSplits...)
	}

	return final
}

func mergeSplits(splits []string, separator string, chunkSize int, chunkOverlap int) []string {
	docs := make([]string, 0)
	currentDoc := make([]string, 0)
	total := 0

	for _, split := range splits {
		splitLen := utf8.RuneCountInString(split)
		sepLen := utf8.RuneCountInString(separator) * compareDocsLen(currentDoc, 0)

		if total+splitLen+sepLen > chunkSize && len(currentDoc) > 0 {
			if doc := strings.TrimSpace(strings.Join(currentDoc, separator)); doc != "" {
				docs = append(docs, doc)
			}

			for len(currentDoc) > 0 && (total > chunkOverlap ||
				(total+splitLen+utf8.RuneCountInString(separator)*compareDocsLen(currentDoc, 1) > chunkSize && total > 0)) {
				total -= utf8.RuneCountInString(currentDoc[0]) + utf8.RuneCountInString(separator)*compareDocsLen(currentDoc, 1)
				currentDoc = currentDoc[1:]
			}
		}

		currentDoc = append(currentDoc, split)
		total += utf8.RuneCountInString(split)
		total += utf8.RuneCountInString(separator) * compareDocsLen(currentDoc, 1)
	}

	if doc := strings.TrimSpace(strings.Join(currentDoc, separator)); doc != "" {
		docs = append(docs, doc)
	}
	return docs
}

func compareDocsLen(currentDocs []string, cmp int) int {
	if len(currentDocs) > cmp {
		return 1
	}
	return 0
}
