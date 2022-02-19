package micropub

type MicroPubPostType struct{
	Name string
	Type string
}

type MicroPubSyndicateTarget struct{
	Name string
	Uid string
}

type MicroPubConfig struct{
	MediaEndpoint string `json:"media-endpoint"`
	PostTypes []MicroPubPostType `json:"post-types"`
	SyndicateTargets []MicroPubSyndicateTarget `json:"syndicate-to"`
}