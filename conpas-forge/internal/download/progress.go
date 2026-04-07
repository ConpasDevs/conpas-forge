package download

import "io"

type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	BytesRead  int64
	OnProgress func(bytesRead, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.BytesRead += int64(n)
	if pr.OnProgress != nil && pr.Total > 0 {
		pr.OnProgress(pr.BytesRead, pr.Total)
	}
	return n, err
}
