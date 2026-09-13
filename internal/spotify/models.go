package spotify

type Paging[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
	Next   any `json:"next,omitempty"`
}

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Artist struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	Genres       []string          `json:"genres,omitempty"`
	Images       []Image           `json:"images,omitempty"`
	ExternalURLs map[string]string `json:"external_urls,omitempty"`
}

type Album struct {
	ID           string              `json:"id"`
	URI          string              `json:"uri"`
	Name         string              `json:"name"`
	AlbumType    string              `json:"album_type,omitempty"`
	ReleaseDate  string              `json:"release_date,omitempty"`
	Artists      []Artist            `json:"artists,omitempty"`
	Images       []Image             `json:"images,omitempty"`
	ExternalURLs map[string]string   `json:"external_urls,omitempty"`
	Items        *Paging[AlbumTrack] `json:"items,omitempty"`
}

type AlbumTrack struct {
	ID         string   `json:"id"`
	URI        string   `json:"uri"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists,omitempty"`
	DurationMS int      `json:"duration_ms,omitempty"`
	Explicit   bool     `json:"explicit"`
}

type Track struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	Artists      []Artist          `json:"artists,omitempty"`
	Album        *Album            `json:"album,omitempty"`
	DurationMS   int               `json:"duration_ms,omitempty"`
	Explicit     bool              `json:"explicit"`
	ExternalURLs map[string]string `json:"external_urls,omitempty"`
}

type Owner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
}

type PlaylistItem struct {
	AddedAt string `json:"added_at,omitempty"`
	Item    Track  `json:"item"`
}

type Playlist struct {
	ID            string                `json:"id"`
	URI           string                `json:"uri"`
	Name          string                `json:"name"`
	Description   string                `json:"description,omitempty"`
	Public        bool                  `json:"public"`
	Collaborative bool                  `json:"collaborative"`
	Owner         Owner                 `json:"owner"`
	Images        []Image               `json:"images,omitempty"`
	ExternalURLs  map[string]string     `json:"external_urls,omitempty"`
	Items         *Paging[PlaylistItem] `json:"items,omitempty"`
}

type Device struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	VolumePercent    int    `json:"volume_percent"`
}

type PlaybackState struct {
	Device       Device `json:"device"`
	RepeatState  string `json:"repeat_state"`
	ShuffleState bool   `json:"shuffle_state"`
	IsPlaying    bool   `json:"is_playing"`
	ProgressMS   int    `json:"progress_ms"`
	Item         *Track `json:"item,omitempty"`
	Context      *struct {
		URI string `json:"uri"`
	} `json:"context,omitempty"`
}

type Queue struct {
	CurrentlyPlaying *Track  `json:"currently_playing,omitempty"`
	Queue            []Track `json:"queue"`
}

type Profile struct {
	ID           string            `json:"id"`
	DisplayName  string            `json:"display_name"`
	URI          string            `json:"uri"`
	ExternalURLs map[string]string `json:"external_urls,omitempty"`
}

type MutationResult struct {
	Completed int `json:"completed"`
}
