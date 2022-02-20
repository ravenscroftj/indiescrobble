package scrobble

const(
	SCROBBLE_TYPE_LISTEN = "listen"
	SCROBBLE_TYPE_TV = "tv"
	SCROBBLE_TYPE_MOVIE = "movie"
	SCROBBLE_TYPE_READ = "read"
)

var ScrobbleTypeEmojis = map[string]string {
	"scrobble" : "🎧",
	"tv" : "📺",
	"movie": "🎬",
	"read": "📖",
};

var ScrobbleTypeNames =  map[string]string {
	"scrobble" : "🎧 Listen",
	"tv" : "📺 TV Show",
	"movie": "🎬 Movie",
	"read": "📖 Read",
};

var ScrobbleTypeVerbs = map[string]string {
	"scrobble" : "listened to",
	"tv" : "watched",
	"movie": "watched",
	"read": "read",
};

var ScrobblePlaceholders =  map[string]string {
	"scrobble" : "Jump Van Halen",
	"tv" : "Schitt's Creek",
	"movie": "Ferris Bueller's Day Off",
	"read": "Three Body Problem Cixin Liu",
};

var ScrobbleCitationProperties = map[string]string {
	"scrobble" : "listen-of",
	"tv" : "watch-of",
	"movie": "watch-of",
	"read": "read-of",
};


type ScrobbleMetaRecord interface{
	GetID() string
	GetDisplayName() string
	GetCanonicalURL() string
	GetThumbnailURL() string
}

type ScrobbleMetaProvider interface{

	GetName() string
	Search(query string) ([]ScrobbleMetaRecord, error)
	GetItem(id string) (ScrobbleMetaRecord, error)
}