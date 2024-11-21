package beat

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	mocks "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()
)

func TestGetBeatFromS3_Success(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	beat := &core.Beat{
		FilePath: "path/to/beat1",
	}
	start, end := 0, 100

	beatStorage.EXPECT().
		GetBeatByID(mock.Anything, 1, core.True).
		Return(beat, nil).
		Once()
	beatStorage.EXPECT().
		GetBeatFromS3(mock.Anything, beat.FilePath, start, &end).
		Return(io.NopCloser(strings.NewReader("Hello, World")), 200, "application/json", nil).
		Once()

	obj, size, contentType, err := beatService.GetBeatFromS3(ctx, 1, start, &end)
	require.NoError(t, err)
	assert.Equal(t, 200, size)
	assert.Equal(t, "application/json", contentType)

	body, err := io.ReadAll(obj)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World", string(body))
}

func TestGetBeatFromS3_Fail(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	s3err := errors.New("s3 error")

	tests := []struct {
		name      string
		behaviour func()
		err       error
	}{
		{
			name: "beat not found",
			behaviour: func() {
				beatStorage.EXPECT().
					GetBeatByID(mock.Anything, 1, core.True).
					Return(nil, core.ErrBeatNotFound).
					Once()
			},
			err: core.ErrBeatNotFound,
		},
		{
			name: "s3 error",
			behaviour: func() {
				beatStorage.EXPECT().
					GetBeatByID(mock.Anything, 1, core.True).
					Return(&core.Beat{}, nil).
					Once()
				beatStorage.EXPECT().
					GetBeatFromS3(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, 0, "", s3err).
					Once()
			},
			err: s3err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.behaviour()

			_, _, _, err := beatService.GetBeatFromS3(ctx, 1, 0, nil)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestWritePartialContent(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	r := io.NopCloser(strings.NewReader("Hello, World"))
	w := &strings.Builder{}
	chunkSize := 5

	err := beatService.WritePartialContent(ctx, r, w, chunkSize)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World", w.String())
}

func TestAddBeat(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	beat := core.Beat{
		ID:   1,
		Name: "beat1",
	}
	beatGenre := []core.BeatGenre{
		{
			Genre: "genre1",
		},
	}
	var path string

	beatStorage.EXPECT().
		AddBeat(mock.Anything, mock.MatchedBy(func(beat core.Beat) bool {
			path = beat.FilePath
			return beat.ID == 1 && beat.Name == "beat1"
		}), beatGenre).
		Return(1, nil).
		Once()

	beatPath, _, err := beatService.AddBeat(ctx, beat, beatGenre)
	require.NoError(t, err)
	assert.Equal(t, path, beatPath)
}

func TestAddBeat_BeatExists(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	beat := core.Beat{
		ID:   1,
		Name: "beat1",
	}
	beatGenre := []core.BeatGenre{
		{
			Genre: "genre1",
		},
	}

	beatStorage.EXPECT().
		AddBeat(mock.Anything, mock.Anything, mock.Anything).
		Return(0, core.ErrBeatExists).
		Once()
	beatStorage.EXPECT().
		GetBeatByID(mock.Anything, 1, core.Any).
		Return(&core.Beat{
			FilePath: "path/to/beat1",
		}, nil).
		Once()

	beatPath, _, err := beatService.AddBeat(ctx, beat, beatGenre)
	require.NoError(t, err)
	assert.Equal(t, "path/to/beat1", beatPath)
}

func TestAddBeat_Fail(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	beat := core.Beat{
		ID:   1,
		Name: "beat1",
	}
	beatGenre := []core.BeatGenre{
		{
			Genre: "genre1",
		},
	}

	beatStorage.EXPECT().
		AddBeat(mock.Anything, mock.Anything, mock.Anything).
		Return(0, errors.New("db error")).
		Once()

	_, _, err := beatService.AddBeat(ctx, beat, beatGenre)
	assert.Error(t, err)
}

func TestGetBeatByFilter(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	userID := 2
	filter := core.FeedFilter{
		Genre: "genre1",
	}

	beatStorage.EXPECT().
		GetUserSeenBeats(mock.Anything, userID).
		Return([]string{"path/to/beat1"}, nil).
		Once()
	beatStorage.EXPECT().
		GetBeatByFilter(mock.Anything, filter, []string{"path/to/beat1"}).
		Return(&core.Beat{
			ID:   1,
			Name: "beat1",
		}, &[]string{"genre1"}[0], nil).
		Once()
	beatStorage.EXPECT().
		PopUserSeenBeat(mock.Anything, userID).
		Return(nil).
		Once()
	beatStorage.EXPECT().
		AddUserSeenBeat(mock.Anything, userID, 1).
		Return(nil).
		Once()

	beat, genre, err := beatService.GetBeatByFilter(ctx, userID, filter)
	require.NoError(t, err)
	assert.Equal(t, "beat1", beat.Name)
	assert.Equal(t, 1, beat.ID)
	assert.Equal(t, "genre1", *genre)
}

func TestGetBeatByFilter_Fail(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	userID := 2
	filter := core.FeedFilter{
		Genre: "genre1",
	}

	beatStorage.EXPECT().
		GetUserSeenBeats(mock.Anything, userID).
		Return([]string{"path/to/beat1"}, nil).
		Once()
	beatStorage.EXPECT().
		GetBeatByFilter(mock.Anything, filter, []string{"path/to/beat1"}).
		Return(nil, nil, errors.New("db error")).
		Once()

	_, _, err := beatService.GetBeatByFilter(ctx, userID, filter)
	assert.Error(t, err)
}

func TestGetBeatByFilter_NotFound(t *testing.T) {
	t.Parallel()

	beatStorage := mocks.NewMockBeatStorage(t)

	beatService := New(beatStorage, 0)

	userID := 2
	filter := core.FeedFilter{
		Genre: "genre1",
	}

	beatStorage.EXPECT().
		GetUserSeenBeats(mock.Anything, userID).
		Return([]string{"path/to/beat1"}, nil).
		Once()
	beatStorage.EXPECT().
		GetBeatByFilter(mock.Anything, filter, []string{"path/to/beat1"}).
		Return(nil, nil, core.ErrBeatNotFound).
		Once()
	beatStorage.EXPECT().
		ClearUserSeenBeats(mock.Anything, userID).
		Return(nil).
		Once()
	beatStorage.EXPECT().
		GetBeatByFilter(mock.Anything, filter, []string{}).
		Return(&core.Beat{
			ID:   1,
			Name: "beat1",
		}, &[]string{"genre1"}[0], nil).
		Once()
	beatStorage.EXPECT().
		PopUserSeenBeat(mock.Anything, userID).
		Return(nil).
		Once()
	beatStorage.EXPECT().
		AddUserSeenBeat(mock.Anything, userID, 1).
		Return(nil).
		Once()

	beat, genre, err := beatService.GetBeatByFilter(ctx, userID, filter)
	require.NoError(t, err)
	assert.Equal(t, "beat1", beat.Name)
	assert.Equal(t, 1, beat.ID)
	assert.Equal(t, "genre1", *genre)
}
