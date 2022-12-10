package shortdescription

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
)

type ShortDescription struct {
	Person      string `json:"person"`
	Description string `json:"description,omitempty"`
}

// apiURL prevents urls from accidentally being used without being processed first.
type apiURL string

const shortDescriptionURL apiURL = "https://en.wikipedia.org/w/api.php?action=query&prop=revisions&rvlimit=1&formatversion=2&format=json&rvprop=content&rvslots=main&titles="

func getShortDescriptionURL(title string) string {
	return string(shortDescriptionURL) + url.QueryEscape(title)
}

func readShortDescription(r io.Reader) (string, error) {
	descr, err := readBetween(r, "{{Short description|", "}}")
	if errors.Is(err, io.EOF) {
		return "", fmt.Errorf("short description %w", ErrNotFound)
	}

	return descr, err
}

// I couldn't find a short implementation that avoided holding data in memory
// while searching the io.Reader, so here's a custom one.
func readBetween(r io.Reader, from, to string) (string, error) {
	br := bufio.NewReader(r)
	if err := discardUntil(br, from); err != nil {
		return "", err
	}

	return readUntil(br, to)
}

func discardUntil(r *bufio.Reader, until string) error {
	_, err := rawReadUntil(r, until, true)
	return err
}

func readUntil(r *bufio.Reader, until string) (string, error) {
	return rawReadUntil(r, until, false)
}

func rawReadUntil(r *bufio.Reader, until string, discard bool) (string, error) {
	var buf []byte

	for {
		chunk, err := r.ReadBytes(until[len(until)-1])
		if err != nil {
			return "", fmt.Errorf("rawReadUntil: %w", err)
		}

		buf = append(buf, chunk...)
		if bytes.HasSuffix(buf, []byte(until)) {
			if discard {
				return "", nil
			}

			return string(buf[:len(buf)-len(until)]), nil
		}

		if discard && len(buf) > len(until) {
			buf = buf[len(buf)-len(until):]
		}
	}
}

// I'm leaving a previous implementation here just in case it's valueble
// for the exercise review. It's shorter but less efficient.

/*var shortDescrRegexp = regexp.MustCompile(`{{Short description\|(.+?)}}`)

func readShortDescription1(r io.Reader) (string, error) {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	indexs := shortDescrRegexp.FindReaderSubmatchIndex(bufio.NewReader(r))
	if len(indexs) < 4 {
		return "", ErrNotFound
	}

	return buf.String()[indexs[2]:indexs[3]], nil
}*/
