package core

type Comics struct {
	ID    int
	URL   string
	Words []string
}

type ComicsInfo struct {
	ID  int
	URL string
}

type IndexRow struct {
	ComicID int
	URL     string
	Word    string
}
