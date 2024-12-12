package beat

import (
	"testing"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBeatByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	t.Run("is_downloaded is false", func(t *testing.T) {
		beat, err := beatStorage.GetBeatByID(ctx, 1, core.Any)
		require.NoError(t, err)

		assert.Equal(t, 101, beat.BeatmakerID)
		assert.Equal(t, "/path/to/beat1", beat.FilePath)
		assert.Equal(t, "hip-hop vibes", beat.Name)
		assert.Equal(t, "a smooth hip-hop beat with relaxing vibes.", beat.Description)
		assert.Equal(t, false, beat.IsFileDownloaded)
		assert.Equal(t, false, beat.IsDeleted)
	})

	t.Run("is_downloaded is true", func(t *testing.T) {
		_, err := beatStorage.GetBeatByID(ctx, 1, core.True)
		assert.ErrorIs(t, err, core.ErrBeatNotFound)
	})
}

func TestAddBeat(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	beat := core.Beat{
		ID:          4,
		BeatmakerID: 1,
		FilePath:    "/path/to/beat4",
		Name:        "synthwave journey",
		Description: "a retro synthwave beat with 80s vibes.",
	}

	beatGenres := []core.BeatGenre{
		{
			BeatID: 4,
			Genre:  "orchestral",
		},
		{
			BeatID: 4,
			Genre:  "pop",
		},
	}

	beatID, err := beatStorage.AddBeat(ctx, beat, beatGenres)
	require.NoError(t, err)

	var gotBeat core.Beat
	err = tdb.DB.
		QueryRow("select id, beatmaker_id, file_path, name, description from beats where id = $1", beatID).
		Scan(
			&gotBeat.ID,
			&gotBeat.BeatmakerID,
			&gotBeat.FilePath,
			&gotBeat.Name,
			&gotBeat.Description)
	require.NoError(t, err)
	assert.Equal(t, beat.ID, gotBeat.ID)
	assert.Equal(t, beat.BeatmakerID, gotBeat.BeatmakerID)
	assert.Equal(t, beat.FilePath, gotBeat.FilePath)
	assert.Equal(t, beat.Name, gotBeat.Name)
	assert.Equal(t, beat.Description, gotBeat.Description)

	rows, err := tdb.DB.
		Query("select beat_id, genre from beats_genres where beat_id = $1 order by id", beatID)
	require.NoError(t, err)

	var cnt int
	for rows.Next() {
		assert.Less(t, cnt, len(beatGenres))

		var gotGenre core.BeatGenre
		err = rows.Scan(&gotGenre.BeatID, &gotGenre.Genre)
		require.NoError(t, err)

		assert.Equal(t, beatGenres[cnt].BeatID, gotGenre.BeatID)
		assert.Equal(t, beatGenres[cnt].Genre, gotGenre.Genre)

		cnt++
	}

	require.NoError(t, rows.Err())
}

func TestGetBeatByFilter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	filter := core.FeedFilter{
		Genres: []string{"lo-fi"},
	}

	beat, genre, err := beatStorage.GetBeatByFilter(ctx, filter, nil)
	require.NoError(t, err)

	assert.Equal(t, 2, beat.ID)
	assert.Equal(t, "lo-fi", *genre)
}

func TestGetBeatByFilter_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	filter := core.FeedFilter{
		Genres: []string{"tr"},
	}

	_, _, err := beatStorage.GetBeatByFilter(ctx, filter, nil)
	assert.ErrorIs(t, err, core.ErrBeatNotFound)
}

func TestGetBeatByFilter_NotFound_Seen(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	filter := core.FeedFilter{
		Genres: []string{"lo"},
	}

	_, _, err := beatStorage.GetBeatByFilter(ctx, filter, []string{"2"})
	assert.ErrorIs(t, err, core.ErrBeatNotFound)
}

func TestGetBeatGenres_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	genres, err := beatStorage.GetBeatGenres(ctx, 100)
	require.NoError(t, err)
	assert.Empty(t, genres)
}

func TestGetBeatGenres_NotEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	beatStorage := New(nil, tdb, "", nil, 0)

	genres, err := beatStorage.GetBeatGenres(ctx, 2)
	require.NoError(t, err)

	assert.Equal(t, 2, len(genres))
	assert.Equal(t, "lo-fi", genres[0].Genre)
	assert.Equal(t, "ambient", genres[1].Genre)
	assert.Equal(t, 2, genres[0].BeatID)
	assert.Equal(t, 2, genres[1].BeatID)
}
