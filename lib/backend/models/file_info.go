package models

type FileInfo struct {
	OriginalName string `json:"original_name"`
	SavedPath    string `json:"saved_path"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
}
