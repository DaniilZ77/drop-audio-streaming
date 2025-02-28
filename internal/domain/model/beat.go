package model

import (
	"time"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	Beat struct {
		ID                  uuid.UUID
		BeatmakerID         uuid.UUID
		ImagePath           string
		Name                string
		Description         string
		IsFileDownloaded    bool
		IsImageDownloaded   bool
		IsArchiveDownloaded bool
		Bpm                 int
		RangeStart          int64
		RangeEnd            int64
		CreatedAt           time.Time
		Genres              []string
		Tags                []string
		Moods               []string
		NoteName            *string
		NoteScale           *string
	}

	GetBeatsNote struct {
		Name  string
		Scale string
	}

	OrderBy struct {
		Order string
		Field string
	}

	GetBeatsParams struct {
		BeatID       *uuid.UUID
		Genre        []string
		Mood         []string
		Tag          []string
		Note         *GetBeatsNote
		BeatmakerID  *uuid.UUID
		BeatName     *string
		Bpm          *int
		OrderBy      *OrderBy
		Limit        uint64
		Offset       uint64
		IsDownloaded *bool
	}

	BeatParams struct {
		Genres []generated.Genre
		Tags   []generated.Tag
		Moods  []generated.Mood
		Notes  []generated.Note
	}

	UpdateBeatParams struct {
		generated.UpdateBeatParams
		Note   *generated.SaveNoteParams
		Genres []generated.SaveGenresParams
		Tags   []generated.SaveTagsParams
		Moods  []generated.SaveMoodsParams
	}

	SaveBeatParams struct {
		generated.SaveBeatParams
		Note   generated.SaveNoteParams
		Genres []generated.SaveGenresParams
		Tags   []generated.SaveTagsParams
		Moods  []generated.SaveMoodsParams
	}

	MediaType string

	AdminScale string

	MediaMeta struct {
		MediaType         MediaType
		HttpContentType   string
		HttpContentLength int64
		Name              string
		Expiry            int64
		UploadURL         string
	}
)

const (
	MediaTypeFile    MediaType  = "file"
	MediaTypeArchive MediaType  = "archive"
	MediaTypeImage   MediaType  = "image"
	AdminScaleMinor  AdminScale = "minor"
	AdminScaleMajor  AdminScale = "major"
)

func ToModelSaveBeatParams(beat *audiov1.UploadBeatRequest) (*SaveBeatParams, error) {
	genres := make([]generated.SaveGenresParams, 0, len(beat.BeatGenre))
	tags := make([]generated.SaveTagsParams, 0, len(beat.BeatTag))
	moods := make([]generated.SaveMoodsParams, 0, len(beat.BeatMood))

	beatID, err := uuid.Parse(beat.BeatId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "beat id must be uuid")
	}

	noteID, err := uuid.Parse(beat.Note.NoteId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "note id must be uuid")
	}

	beatmakerID, err := uuid.Parse(beat.BeatmakerId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "beatmaker id must be uuid")
	}

	var id uuid.UUID
	for _, v := range beat.BeatGenre {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "genre id must be uuid")
		}
		genres = append(genres, generated.SaveGenresParams{
			GenreID: id,
			BeatID:  beatID,
		})
	}

	for _, v := range beat.BeatTag {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "tag id must be uuid")
		}
		tags = append(tags, generated.SaveTagsParams{
			TagID:  id,
			BeatID: beatID,
		})
	}

	for _, v := range beat.BeatMood {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "mood id must be uuid")
		}
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: id,
			BeatID: beatID,
		})
	}

	return &SaveBeatParams{
		SaveBeatParams: generated.SaveBeatParams{
			ID:          beatID,
			BeatmakerID: beatmakerID,
			Name:        beat.Name,
			Description: beat.Description,
			Bpm:         int32(beat.Bpm),
			RangeStart:  beat.Range.Start,
			RangeEnd:    beat.Range.End,
		},
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Note: generated.SaveNoteParams{
			BeatID: beatID,
			NoteID: noteID,
			Scale:  generated.NoteScale(beat.Note.Scale),
		},
	}, nil
}

func ToModelGetBeatsParams(params *audiov1.GetBeatsRequest) (*GetBeatsParams, error) {
	var res GetBeatsParams

	var id uuid.UUID
	var err error
	if params.BeatId != nil {
		if id, err = uuid.Parse(*params.BeatId); err != nil {
			return nil, NewErr(ErrInvalidID, "beat id must be uuid")
		}
		beatID := id
		res.BeatID = &beatID
	}

	res.Genre = params.Genre
	res.Mood = params.Mood
	res.Tag = params.Tag

	if params.Note != nil {
		res.Note = &GetBeatsNote{
			Name:  params.Note.Name,
			Scale: params.Note.Scale,
		}
	}

	if params.BeatmakerId != nil {
		if id, err = uuid.Parse(*params.BeatmakerId); err != nil {
			return nil, NewErr(ErrInvalidID, "beatmaker id must be uuid")
		}
		beatmakerID := id
		res.BeatmakerID = &beatmakerID
	}

	res.BeatName = params.BeatName

	if params.Bpm != nil {
		bpm := int(*params.Bpm)
		res.Bpm = &bpm
	}

	if params.OrderBy != nil {
		res.OrderBy = &OrderBy{
			Order: params.OrderBy.Order,
			Field: params.OrderBy.Field,
		}
	}

	res.Limit = params.Limit
	res.Offset = params.Offset

	res.IsDownloaded = params.IsDownloaded

	return &res, nil
}

func ToGetBeatsResponse(beats []Beat, users []*userv1.GetUserResponse, total uint64, params GetBeatsParams) *audiov1.GetBeatsResponse {
	var res []*audiov1.Beat
	for i := range beats {
		res = append(res, toResponseBeat(beats[i], users[i]))
	}

	return &audiov1.GetBeatsResponse{
		Beats: res,
		Pagination: &audiov1.Pagination{
			Records:        total,
			RecordsPerPage: params.Limit,
			Pages:          (total + params.Limit - 1) / params.Limit,
			CurPage:        params.Offset/params.Limit + 1,
		},
	}
}

func toResponseBeat(b Beat, u *userv1.GetUserResponse) *audiov1.Beat {
	scale := generated.NoteScaleMinor
	if b.NoteScale != nil && *b.NoteScale == string(generated.NoteScaleMajor) {
		scale = generated.NoteScaleMajor
	}

	var name string
	if b.NoteName != nil {
		name = *b.NoteName
	}

	return &audiov1.Beat{
		BeatId: b.ID.String(),
		Beatmaker: &audiov1.Beatmaker{
			Id:        b.BeatmakerID.String(),
			Username:  u.Username,
			Pseudonym: u.Pseudonym,
		},
		ImageDownloadUrl: b.ImagePath,
		Name:             b.Name,
		Description:      b.Description,
		Genre:            b.Genres,
		Tag:              b.Tags,
		Mood:             b.Moods,
		Note: &audiov1.GetBeatsNote{
			Name:  name,
			Scale: string(scale),
		},
		Bpm: int64(b.Bpm),
		Range: &audiov1.Range{
			Start: b.RangeStart,
			End:   b.RangeEnd,
		},
		IsFileUploaded:    b.IsFileDownloaded,
		IsImageUploaded:   b.IsImageDownloaded,
		IsArchiveUploaded: b.IsArchiveDownloaded,
		CreatedAt:         timestamppb.New(b.CreatedAt),
	}
}

func ToGetBeatParamsResponse(b BeatParams) *audiov1.GetBeatParamsResponse {
	genres := make([]*audiov1.GenreParam, 0, len(b.Genres))
	tags := make([]*audiov1.TagParam, 0, len(b.Tags))
	moods := make([]*audiov1.MoodParam, 0, len(b.Moods))
	notes := make([]*audiov1.NoteParam, 0, len(b.Notes))

	for _, v := range b.Genres {
		genres = append(genres, &audiov1.GenreParam{
			GenreId: v.ID.String(),
			Name:    v.Name,
		})
	}

	for _, v := range b.Tags {
		tags = append(tags, &audiov1.TagParam{
			TagId: v.ID.String(),
			Name:  v.Name,
		})
	}

	for _, v := range b.Moods {
		moods = append(moods, &audiov1.MoodParam{
			MoodId: v.ID.String(),
			Name:   v.Name,
		})
	}

	for _, v := range b.Notes {
		notes = append(notes, &audiov1.NoteParam{
			NoteId: v.ID.String(),
			Name:   v.Name,
		})
	}

	return &audiov1.GetBeatParamsResponse{
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Notes:  notes,
	}
}

func ToModelUpdateBeatParams(req *audiov1.UpdateBeatRequest) (*UpdateBeatParams, error) {
	var genres []generated.SaveGenresParams
	var tags []generated.SaveTagsParams
	var moods []generated.SaveMoodsParams

	beatID, err := uuid.Parse(req.BeatId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "beat id must be uuid")
	}

	var id uuid.UUID
	for _, v := range req.BeatGenre {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "genre id must be uuid")
		}
		genres = append(genres, generated.SaveGenresParams{
			GenreID: id,
			BeatID:  beatID,
		})
	}

	for _, v := range req.BeatTag {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "tag id must be uuid")
		}
		tags = append(tags, generated.SaveTagsParams{
			TagID:  id,
			BeatID: beatID,
		})
	}

	for _, v := range req.BeatMood {
		if id, err = uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "mood id must be uuid")
		}
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: id,
			BeatID: beatID,
		})
	}

	var note *generated.SaveNoteParams
	if req.Note != nil {
		if id, err = uuid.Parse(req.Note.NoteId); err != nil {
			return nil, NewErr(ErrInvalidID, "note id must be uuid")
		}
		note = &generated.SaveNoteParams{
			BeatID: beatID,
			NoteID: id,
			Scale:  generated.NoteScale(req.Note.Scale),
		}
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var bpm *int32
	if req.Bpm != nil {
		tmp := int32(*req.Bpm)
		bpm = &tmp
	}

	var isImageDownloaded *bool
	if req.UpdateImage != nil && *req.UpdateImage {
		tmp := false
		isImageDownloaded = &tmp
	}

	var isFileDownloaded *bool
	if req.UpdateFile != nil && *req.UpdateFile {
		tmp := false
		isFileDownloaded = &tmp
	}

	var isArchiveDownloaded *bool
	if req.UpdateArchive != nil && *req.UpdateArchive {
		tmp := false
		isArchiveDownloaded = &tmp
	}

	var rangeStart, rangeEnd *int64
	if req.Range != nil {
		rangeStart = &req.Range.Start
		rangeEnd = &req.Range.End
	}

	return &UpdateBeatParams{
		UpdateBeatParams: generated.UpdateBeatParams{
			ID:                  beatID,
			Name:                name,
			Description:         description,
			Bpm:                 bpm,
			IsImageDownloaded:   isImageDownloaded,
			IsFileDownloaded:    isFileDownloaded,
			IsArchiveDownloaded: isArchiveDownloaded,
			RangeStart:          rangeStart,
			RangeEnd:            rangeEnd,
		},
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Note:   note,
	}, nil
}

func ToModelGetArchiveParams(req *audiov1.AcquireBeatRequest) (*generated.SaveOwnerParams, error) {
	var beatID, userID uuid.UUID
	var err error

	if beatID, err = uuid.Parse(req.BeatId); err != nil {
		return nil, NewErr(ErrInvalidID, "beat id must be uuid")
	}

	if userID, err = uuid.Parse(req.UserId); err != nil {
		return nil, NewErr(ErrInvalidID, "user id must be uuid")
	}

	return &generated.SaveOwnerParams{
		BeatID: beatID,
		UserID: userID,
	}, nil
}
