package opds

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestEntryWithDateOnlyPublished(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<entry xmlns="http://www.w3.org/2005/Atom">
	<title>Test Book</title>
	<id>test-id</id>
	<published>2006-07-18</published>
</entry>`

	var entry Entry
	err := xml.NewDecoder(strings.NewReader(xmlData)).Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	if entry.Published == nil {
		t.Error("Expected published date to be parsed")
		return
	}

	expected := time.Date(2006, 7, 18, 0, 0, 0, 0, time.UTC)
	if !entry.Published.Time.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, entry.Published.Time)
	}
}

func TestEntryUnmarshal(t *testing.T) {
	xmlData := `
	 <entry>
		<title>Pride and Prejudice</title>
		<author>
		<name>Jane Austen</name>
		</author>
		<id>urn:uuid:07c4b45f-dc10-4c1e-937b-1f994b67e921</id>
		<updated>2021-04-07T16:41:14+00:00</updated>
		<published>2019-10-23T23:17:35+00:00</published>
		<dc:date>2014-05-06T04:00:00+00:00</dc:date>
		<content type="xhtml">
		<div xmlns="http://www.w3.org/1999/xhtml">RATING: ★★★★<br/>
	Formats: AZW3,EPUB<br/>
	<div><p>A timeless romance following Elizabeth Bennet, a strong-willed young woman, and Mr. Darcy, a proud and wealthy gentleman. Set in Georgian England, the novel explores themes of love, marriage, social class, and personal growth through wit and humor.</p><p>When Elizabeth first meets Mr. Darcy at a ball, she finds him arrogant and disagreeable. Meanwhile, she is charmed by the dashing Mr. Wickham, who tells her tales of Darcy's alleged misconduct. As the story unfolds, Elizabeth discovers that first impressions can be deceiving, and that pride and prejudice can blind us to true character.</p><p>Through a series of misunderstandings, revelations, and personal growth, both Elizabeth and Darcy must overcome their initial judgments to find true love. This beloved classic remains one of the most popular novels in English literature.</p></div></div>
		</content>
		<link type="application/x-mobi8-ebook" href="/get/azw3/313/books" rel="http://opds-spec.org/acquisition" length="825549" mtime="2021-04-08T00:38:35+00:00"/>
		<link type="application/epub+zip" href="/get/epub/313/books" rel="http://opds-spec.org/acquisition" length="642930" mtime="2021-04-08T00:38:23+00:00"/>
		<link type="image/jpeg" href="/get/cover/313/books" rel="http://opds-spec.org/cover"/>
		<link type="image/jpeg" href="/get/thumb/313/books" rel="http://opds-spec.org/thumbnail"/>
		<link type="image/jpeg" href="/get/cover/313/books" rel="http://opds-spec.org/image"/>
		<link type="image/jpeg" href="/get/thumb/313/books" rel="http://opds-spec.org/image/thumbnail"/>
	</entry>
  <entry>`

	var entry Entry
	err := xml.NewDecoder(strings.NewReader(xmlData)).Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	// ensure we parse the content correctly
	if strings.TrimSpace(entry.Content.Content) == "" {
		t.Error("Expected content to be parsed, got empty string")
	}

	// Verify content contains expected text
	if !strings.Contains(entry.Content.Content, "RATING: ★★★★") {
		t.Error("Expected content to contain rating, but it was not found")
	}

	if !strings.Contains(entry.Content.Content, "Elizabeth Bennet") {
		t.Error("Expected content to contain main character name, but it was not found")
	}

	if entry.Title != "Pride and Prejudice" {
		t.Errorf("Expected title 'Pride and Prejudice', got '%s'", entry.Title)
	}
}