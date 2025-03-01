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
		Bpm                 int64
		RangeStart          int64
		RangeEnd            int64
		CreatedAt           time.Time
		Genres              []string
		Tags                []string
		Moods               []string
		NoteName            *string
		NoteScale           *string
	}

	BeatsNote struct {
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
		Note         *BeatsNote
		BeatmakerID  *uuid.UUID
		BeatName     *string
		Bpm          *int64
		OrderBy      *OrderBy
		Limit        uint64
		Offset       uint64
		IsDownloaded *bool
	}

	BeatAttributes struct {
		Genres []generated.Genre
		Tags   []generated.Tag
		Moods  []generated.Mood
		Notes  []generated.Note
	}

	UpdateBeat struct {
		generated.UpdateBeatParams
		Note   *generated.SaveNoteParams
		Genres []generated.SaveGenresParams
		Tags   []generated.SaveTagsParams
		Moods  []generated.SaveMoodsParams
	}

	SaveBeat struct {
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

func toDomainSaveGenresParams(beatID uuid.UUID, genres []string) ([]generated.SaveGenresParams, error) {
	var res []generated.SaveGenresParams
	for _, v := range genres {
		if genreID, err := uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "genre id must be uuid")
		} else {
			res = append(res, generated.SaveGenresParams{BeatID: beatID, GenreID: genreID})
		}
	}
	return res, nil
}

func toDomainSaveTagsParams(beatID uuid.UUID, tags []string) ([]generated.SaveTagsParams, error) {
	var res []generated.SaveTagsParams
	for _, v := range tags {
		if tagID, err := uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "tag id must be uuid")
		} else {
			res = append(res, generated.SaveTagsParams{BeatID: beatID, TagID: tagID})
		}
	}
	return res, nil
}

func toDomainSaveMoodsParams(beatID uuid.UUID, moods []string) ([]generated.SaveMoodsParams, error) {
	var res []generated.SaveMoodsParams
	for _, v := range moods {
		if moodID, err := uuid.Parse(v); err != nil {
			return nil, NewErr(ErrInvalidID, "mood id must be uuid")
		} else {
			res = append(res, generated.SaveMoodsParams{BeatID: beatID, MoodID: moodID})
		}
	}
	return res, nil
}

func ToDomainSaveBeat(beat *audiov1.UploadBeatRequest) (*SaveBeat, error) {
	var res SaveBeat
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
	res.ID = beatID
	res.BeatmakerID = beatmakerID
	res.Name = beat.Name
	res.Description = beat.Description
	res.Bpm = int32(beat.Bpm)
	res.RangeStart = beat.Range.Start
	res.RangeEnd = beat.Range.End
	res.Note.BeatID = beatID
	res.Note.NoteID = noteID
	res.Note.Scale = generated.NoteScale(beat.Note.Scale)
	res.Genres, err = toDomainSaveGenresParams(beatID, beat.BeatGenre)
	if err != nil {
		return nil, err
	}
	res.Tags, err = toDomainSaveTagsParams(beatID, beat.BeatTag)
	if err != nil {
		return nil, err
	}
	res.Moods, err = toDomainSaveMoodsParams(beatID, beat.BeatMood)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func ToDomainGetBeatsParams(params *audiov1.GetBeatsRequest) (*GetBeatsParams, error) {
	var res GetBeatsParams
	if params.BeatId != nil {
		if beatID, err := uuid.Parse(*params.BeatId); err != nil {
			return nil, NewErr(ErrInvalidID, "beat id must be uuid")
		} else {
			res.BeatID = &beatID
		}
	}
	res.Genre = params.Genre
	res.Mood = params.Mood
	res.Tag = params.Tag
	if params.Note != nil {
		res.Note = &BeatsNote{
			Name:  params.Note.Name,
			Scale: params.Note.Scale,
		}
	}
	if params.BeatmakerId != nil {
		if beatmakerID, err := uuid.Parse(*params.BeatmakerId); err != nil {
			return nil, NewErr(ErrInvalidID, "beatmaker id must be uuid")
		} else {
			res.BeatmakerID = &beatmakerID
		}
	}
	res.BeatName = params.BeatName
	res.Bpm = params.Bpm
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
	var res audiov1.GetBeatsResponse
	for i := range beats {
		res.Beats = append(res.Beats, toResponseBeat(beats[i], users[i]))
	}
	res.Pagination = &audiov1.Pagination{}
	res.Pagination.Records = total
	res.Pagination.RecordsPerPage = params.Limit
	res.Pagination.Pages = (total + params.Limit - 1) / params.Limit
	res.Pagination.CurPage = params.Offset/params.Limit + 1
	return &res
}

func toResponseBeat(b Beat, u *userv1.GetUserResponse) *audiov1.Beat {
	var res audiov1.Beat
	res.Note = &audiov1.GetBeatsNote{}
	res.Note.Scale = string(generated.NoteScaleMinor)
	if b.NoteScale != nil && *b.NoteScale == string(generated.NoteScaleMajor) {
		res.Note.Scale = string(generated.NoteScaleMajor)
	}
	if b.NoteName != nil {
		res.Note.Name = *b.NoteName
	}
	res.BeatId = b.ID.String()
	res.Beatmaker = &audiov1.Beatmaker{}
	res.Beatmaker.Id = b.BeatmakerID.String()
	res.Beatmaker.Pseudonym = u.Pseudonym
	res.Beatmaker.Username = u.Username
	res.ImageDownloadUrl = b.ImagePath
	res.Name = b.Name
	res.Description = b.Description
	res.Genre = b.Genres
	res.Tag = b.Tags
	res.Mood = b.Moods
	res.Bpm = b.Bpm
	res.Range = &audiov1.Range{}
	res.Range.Start = b.RangeStart
	res.Range.End = b.RangeEnd
	res.IsFileUploaded = b.IsFileDownloaded
	res.IsImageUploaded = b.IsImageDownloaded
	res.IsArchiveUploaded = b.IsArchiveDownloaded
	res.CreatedAt = timestamppb.New(b.CreatedAt)
	return &res
}

func ToGetBeatParamsResponse(b BeatAttributes) *audiov1.GetBeatParamsResponse {
	var res audiov1.GetBeatParamsResponse
	for _, v := range b.Genres {
		res.Genres = append(res.Genres, &audiov1.GenreParam{GenreId: v.ID.String(), Name: v.Name})
	}
	for _, v := range b.Tags {
		res.Tags = append(res.Tags, &audiov1.TagParam{TagId: v.ID.String(), Name: v.Name})
	}
	for _, v := range b.Moods {
		res.Moods = append(res.Moods, &audiov1.MoodParam{MoodId: v.ID.String(), Name: v.Name})
	}
	for _, v := range b.Notes {
		res.Notes = append(res.Notes, &audiov1.NoteParam{NoteId: v.ID.String(), Name: v.Name})
	}
	return &res
}

func ToDomainUpdateBeat(req *audiov1.UpdateBeatRequest) (*UpdateBeat, error) {
	var res UpdateBeat
	beatID, err := uuid.Parse(req.BeatId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "beat id must be uuid")
	}
	res.ID = beatID
	res.Genres, err = toDomainSaveGenresParams(beatID, req.BeatGenre)
	if err != nil {
		return nil, err
	}
	res.Tags, err = toDomainSaveTagsParams(beatID, req.BeatTag)
	if err != nil {
		return nil, err
	}
	res.Moods, err = toDomainSaveMoodsParams(beatID, req.BeatMood)
	if err != nil {
		return nil, err
	}
	if req.Note != nil {
		if noteID, err := uuid.Parse(req.Note.NoteId); err != nil {
			return nil, NewErr(ErrInvalidID, "note id must be uuid")
		} else {
			res.Note = &generated.SaveNoteParams{}
			res.Note.BeatID = beatID
			res.Note.NoteID = noteID
			res.Note.Scale = generated.NoteScale(req.Note.Scale)
		}
	}
	res.Name = req.Name
	res.Description = req.Description
	if req.Bpm != nil {
		tmp := int32(*req.Bpm)
		res.Bpm = &tmp
	}
	if req.UpdateImage != nil && *req.UpdateImage {
		tmp := false
		res.IsImageDownloaded = &tmp
	}
	if req.UpdateFile != nil && *req.UpdateFile {
		tmp := false
		res.IsFileDownloaded = &tmp
	}
	if req.UpdateArchive != nil && *req.UpdateArchive {
		tmp := false
		res.IsArchiveDownloaded = &tmp
	}
	if req.Range != nil {
		res.RangeStart = &req.Range.Start
		res.RangeEnd = &req.Range.End
	}
	return &res, nil
}

func ToDomainSaveOwnerParams(req *audiov1.AcquireBeatRequest) (*generated.SaveOwnerParams, error) {
	var res generated.SaveOwnerParams
	beatID, err := uuid.Parse(req.BeatId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "beat id must be uuid")
	}
	res.BeatID = beatID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, NewErr(ErrInvalidID, "user id must be uuid")
	}
	res.UserID = userID
	return &res, nil
}
