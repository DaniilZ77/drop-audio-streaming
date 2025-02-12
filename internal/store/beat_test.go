package beat

// import (
// 	"testing"

// 	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestGetBeatByID(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Parallel()

// 	beatStorage := New(nil, tdb, "", nil, 0)

// 	t.Run("is_downloaded is false", func(t *testing.T) {
// 		beat, err := beatStorage.GetBeatByID(ctx, 1, core.Any)
// 		require.NoError(t, err)

// 		assert.Equal(t, 101, beat.BeatmakerID)
// 		assert.Equal(t, "/path/to/beat1", beat.FilePath)
// 		assert.Equal(t, "hip-hop vibes", beat.Name)
// 		assert.Equal(t, "a smooth hip-hop beat with relaxing vibes.", beat.Description)
// 		assert.Equal(t, false, beat.IsFileDownloaded)
// 		assert.Equal(t, false, beat.IsDeleted)
// 	})

// 	t.Run("is_downloaded is true", func(t *testing.T) {
// 		_, err := beatStorage.GetBeatByID(ctx, 1, core.True)
// 		assert.ErrorIs(t, err, core.ErrBeatNotFound)
// 	})
// }

// func TestAddBeat(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Parallel()

// 	beatStorage := New(nil, tdb, "", nil, 0)

// 	beat := core.BeatParams{
// 		Beat: core.Beat{
// 			ID:          4,
// 			BeatmakerID: 1,
// 			FilePath:    "/path/to/beat4",
// 			Name:        "synthwave journey",
// 			Description: "a retro synthwave beat with 80s vibes.",
// 		},
// 		Genres: []core.BeatGenre{
// 			{
// 				BeatID:  4,
// 				GenreID: 1,
// 			},
// 			{
// 				BeatID:  4,
// 				GenreID: 2,
// 			},
// 		},
// 		Tags: []core.BeatTag{
// 			{
// 				BeatID: 4,
// 				TagID:  1,
// 			},
// 		},
// 		Moods: []core.BeatMood{
// 			{
// 				BeatID: 4,
// 				MoodID: 1,
// 			},
// 		},
// 		Note: core.BeatNote{
// 			BeatID: 4,
// 			NoteID: 1,
// 			Scale:  "minor",
// 		},
// 	}

// 	beatID, err := beatStorage.AddBeat(ctx, beat)
// 	require.NoError(t, err)

// 	var gotBeat core.Beat
// 	err = tdb.DB.
// 		QueryRow("select id, beatmaker_id, file_path, name, description from beats where id = $1", beatID).
// 		Scan(
// 			&gotBeat.ID,
// 			&gotBeat.BeatmakerID,
// 			&gotBeat.FilePath,
// 			&gotBeat.Name,
// 			&gotBeat.Description)
// 	require.NoError(t, err)
// 	assert.Equal(t, beat.Beat.ID, gotBeat.ID)
// 	assert.Equal(t, beat.Beat.BeatmakerID, gotBeat.BeatmakerID)
// 	assert.Equal(t, beat.Beat.FilePath, gotBeat.FilePath)
// 	assert.Equal(t, beat.Beat.Name, gotBeat.Name)
// 	assert.Equal(t, beat.Beat.Description, gotBeat.Description)

// 	rows, err := tdb.DB.
// 		Query("select beat_id, genre_id from beats_genres where beat_id = $1 order by id", beatID)
// 	require.NoError(t, err)

// 	var cnt int
// 	for rows.Next() {
// 		assert.Less(t, cnt, len(beat.Genres))

// 		var gotGenre core.BeatGenre
// 		err = rows.Scan(&gotGenre.BeatID, &gotGenre.GenreID)
// 		require.NoError(t, err)

// 		assert.Equal(t, beat.Genres[cnt].BeatID, gotGenre.BeatID)
// 		assert.Equal(t, beat.Genres[cnt].GenreID, gotGenre.GenreID)

// 		cnt++
// 	}

// 	require.NoError(t, rows.Err())
// }

// func TestGetBeatByFilter(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Parallel()

// 	beatStorage := New(nil, tdb, "", nil, 0)

// 	filter := core.FeedFilter{
// 		Genres: []int{3},
// 	}

// 	beat, err := beatStorage.GetBeatByFilter(ctx, filter, nil)
// 	require.NoError(t, err)

// 	assert.Equal(t, 3, beat.Beat.ID)
// 	require.GreaterOrEqual(t, len(beat.Genres), 1)
// 	assert.Equal(t, 3, beat.Genres[0].GenreID)
// }

// func TestGetBeatByFilter_NotFound(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Parallel()

// 	beatStorage := New(nil, tdb, "", nil, 0)

// 	filter := core.FeedFilter{
// 		Genres: []int{10},
// 	}

// 	_, err := beatStorage.GetBeatByFilter(ctx, filter, nil)
// 	assert.ErrorIs(t, err, core.ErrBeatNotFound)
// }

// func TestGetBeatByFilter_NotFound_Seen(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Parallel()

// 	beatStorage := New(nil, tdb, "", nil, 0)

// 	filter := core.FeedFilter{
// 		Genres: []int{4},
// 	}

// 	_, err := beatStorage.GetBeatByFilter(ctx, filter, []string{"1"})
// 	assert.ErrorIs(t, err, core.ErrBeatNotFound)
// }
