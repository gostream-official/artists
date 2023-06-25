package models

// Description:
//
//	The data model definition for an artist.
//	This is a direct reference to the database data model.
type ArtistInfo struct {

	// The id of the artist (primary key).
	ID string `json:"id" bson:"_id"`

	// The name of the artist.
	Name string `json:"name" bson:"name"`

	// The genres an artist is active in.
	Genres []string `json:"genres" bson:"genres"`

	// The amount of followers the artist has.
	Followers uint32 `json:"followers" bson:"followers"`

	// Some artist statistics.
	Stats ArtistStats `json:"stats" bson:"stats"`
}

// Description:
//
//	Some artist statistics.
type ArtistStats struct {

	// The popularity factor of the artist.
	Popularity float32 `json:"popularity" bson:"popularity"`
}
