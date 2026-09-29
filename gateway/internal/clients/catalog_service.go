package clients

import (
	"context"

	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Track struct {
	TrackUUID  string `json:"track_uuid"`
	TrackName  string `json:"track_name"`
	ArtistUUID string `json:"artist_uuid"`
	ArtistName string `json:"artist_name"`
	AlbumUUID  string `json:"album_uuid"`
	AlbumName  string `json:"album_name"`
	Explicit   bool   `json:"explicit"`
	DurationMs int32  `json:"duration_ms"`
	CoverPath  string `json:"cover_path"`
}

type LightTrack struct {
	TrackUUID  string `json:"track_uuid"`
	TrackName  string `json:"track_name"`
	Explicit   bool   `json:"explicit"`
	DurationMs int32  `json:"duration_ms"`
}

type TrackPath struct {
	Path       string `json:"path"`
	DurationMs int32  `json:"duration_ms"`
	AlbumUUID  string `json:"album_uuid"`
	CoverPath  string `json:"cover_path"`
}

type Artist struct {
	ArtistUUID string `json:"artist_uuid"`
	ArtistName string `json:"artist_name"`
}

type LightAlbum struct {
	AlbumUUID string `json:"album_uuid"`
	AlbumName string `json:"album_name"`
	CreatedAt string `json:"created_at"`
}

type Album struct {
	AlbumUUID  string `json:"album_uuid"`
	ArtistUUID string `json:"artist_uuid"`
	AlbumName  string `json:"album_name"`
	CreatedAt  string `json:"created_at"`
	CoverPath  string `json:"cover_path"`
}

type TrackDetails struct {
	TrackUUID  string `json:"track_uuid"`
	TrackName  string `json:"track_name"`
	ArtistUUID string `json:"artist_uuid"`
	AlbumUUID  string `json:"album_uuid"`
	Explicit   bool   `json:"explicit"`
	Path       string `json:"path"`
	DurationMs int32  `json:"duration_ms"`
}

type AlbumTrackInput struct {
	TrackUUID string `json:"track_uuid"`
	Position  int32  `json:"position"`
}

type CatalogClient struct {
	conn            *grpc.ClientConn
	writeConn       *grpc.ClientConn
	grpcClient      catalogv1.CatalogServiceClient
	grpcWriteClient catalogv1.CatalogWriteServiceClient
}

func NewCatalogClient(address, writeAddress string) (*CatalogClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	writeConn, err := grpc.NewClient(writeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CatalogClient{
		conn:            conn,
		writeConn:       writeConn,
		grpcClient:      catalogv1.NewCatalogServiceClient(conn),
		grpcWriteClient: catalogv1.NewCatalogWriteServiceClient(writeConn),
	}, nil
}

func (c *CatalogClient) GetTrackByID(ctx context.Context, trackUUID string) (*TrackPath, error) {
	res, err := c.grpcClient.GetTrackByID(ctx, &catalogv1.GetTrackByIDRequest{TrackUuid: trackUUID})
	if err != nil {
		return nil, err
	}
	return &TrackPath{Path: res.Path, DurationMs: res.DurationMs, AlbumUUID: res.AlbumUuid, CoverPath: res.CoverPath}, nil
}

func (c *CatalogClient) TrackExists(ctx context.Context, trackUUID string) (bool, error) {
	res, err := c.grpcClient.TrackExists(ctx, &catalogv1.TrackExistsRequest{TrackUuid: trackUUID})
	if err != nil {
		return false, err
	}
	return res.Exists, nil
}

func (c *CatalogClient) SearchTrack(ctx context.Context, trackName string, limit, offset int32) ([]Track, error) {
	res, err := c.grpcClient.SearchTrack(ctx, &catalogv1.SearchTrackRequest{TrackName: trackName, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	tracks := make([]Track, 0, len(res.Tracks))
	for _, t := range res.Tracks {
		tracks = append(tracks, Track{
			TrackUUID:  t.TrackUuid,
			TrackName:  t.TrackName,
			ArtistUUID: t.ArtistUuid,
			ArtistName: t.ArtistName,
			AlbumUUID:  t.AlbumUuid,
			AlbumName:  t.AlbumName,
			Explicit:   t.Explicit,
			DurationMs: t.DurationMs,
			CoverPath:  t.CoverPath,
		})
	}
	return tracks, nil
}

func (c *CatalogClient) GetArtistByID(ctx context.Context, artistUUID string) (*Artist, error) {
	res, err := c.grpcClient.GetArtistByID(ctx, &catalogv1.GetArtistByIDRequest{ArtistUuid: artistUUID})
	if err != nil {
		return nil, err
	}
	return &Artist{ArtistUUID: res.Artist.ArtistUuid, ArtistName: res.Artist.ArtistName}, nil
}

func (c *CatalogClient) SearchArtist(ctx context.Context, artistName string, limit, offset int32) ([]Artist, error) {
	res, err := c.grpcClient.SearchArtist(ctx, &catalogv1.SearchArtistRequest{ArtistName: artistName, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	artists := make([]Artist, 0, len(res.Artists))
	for _, a := range res.Artists {
		artists = append(artists, Artist{ArtistUUID: a.ArtistUuid, ArtistName: a.ArtistName})
	}
	return artists, nil
}

func (c *CatalogClient) GetArtistAlbums(ctx context.Context, artistUUID string, limit, offset int32) ([]LightAlbum, error) {
	res, err := c.grpcClient.GetArtistAlbums(ctx, &catalogv1.GetArtistAlbumsRequest{ArtistUuid: artistUUID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	albums := make([]LightAlbum, 0, len(res.ArtistAlbums))
	for _, a := range res.ArtistAlbums {
		albums = append(albums, LightAlbum{AlbumUUID: a.AlbumUuid, AlbumName: a.AlbumName, CreatedAt: a.CreatedAt})
	}
	return albums, nil
}

func (c *CatalogClient) GetArtistTracks(ctx context.Context, artistUUID string, limit, offset int32) ([]LightTrack, error) {
	res, err := c.grpcClient.GetArtistTracks(ctx, &catalogv1.GetArtistTracksRequest{ArtistUuid: artistUUID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	tracks := make([]LightTrack, 0, len(res.ArtistTracks))
	for _, t := range res.ArtistTracks {
		tracks = append(tracks, LightTrack{TrackUUID: t.TrackUuid, TrackName: t.TrackName, Explicit: t.Explicit, DurationMs: t.DurationMs})
	}
	return tracks, nil
}

func (c *CatalogClient) GetAlbumByID(ctx context.Context, albumUUID string) (*Album, error) {
	res, err := c.grpcClient.GetAlbumByID(ctx, &catalogv1.GetAlbumByIDRequest{AlbumUuid: albumUUID})
	if err != nil {
		return nil, err
	}
	return &Album{
		AlbumUUID:  res.Album.AlbumUuid,
		ArtistUUID: res.Album.ArtistUuid,
		AlbumName:  res.Album.AlbumName,
		CreatedAt:  res.Album.CreatedAt,
		CoverPath:  res.Album.CoverPath,
	}, nil
}

func (c *CatalogClient) GetAlbumTracks(ctx context.Context, albumUUID string) ([]LightTrack, error) {
	res, err := c.grpcClient.GetAlbumTracks(ctx, &catalogv1.GetAlbumTracksRequest{AlbumUuid: albumUUID})
	if err != nil {
		return nil, err
	}
	tracks := make([]LightTrack, 0, len(res.LightTrack))
	for _, t := range res.LightTrack {
		tracks = append(tracks, LightTrack{TrackUUID: t.TrackUuid, TrackName: t.TrackName, Explicit: t.Explicit, DurationMs: t.DurationMs})
	}
	return tracks, nil
}

func (c *CatalogClient) AddArtist(ctx context.Context, artistName string) (*Artist, error) {
	res, err := c.grpcWriteClient.AddArtist(ctx, &catalogv1.AddArtistRequest{ArtistName: artistName})
	if err != nil {
		return nil, err
	}
	return &Artist{ArtistUUID: res.Artist.ArtistUuid, ArtistName: res.Artist.ArtistName}, nil
}

func (c *CatalogClient) AddAlbum(ctx context.Context, artistUUID, albumName string) (*Album, error) {
	res, err := c.grpcWriteClient.AddAlbum(ctx, &catalogv1.AddAlbumRequest{ArtistUuid: artistUUID, AlbumName: albumName})
	if err != nil {
		return nil, err
	}
	return &Album{
		AlbumUUID:  res.Album.AlbumUuid,
		ArtistUUID: res.Album.ArtistUuid,
		AlbumName:  res.Album.AlbumName,
		CreatedAt:  res.Album.CreatedAt,
		CoverPath:  res.Album.CoverPath,
	}, nil
}

func (c *CatalogClient) AddTrack(ctx context.Context, trackName, artistUUID, albumUUID string, explicit bool, path string, durationMs int32) (*TrackDetails, error) {
	res, err := c.grpcWriteClient.AddTrack(ctx, &catalogv1.AddTrackRequest{
		TrackName:  trackName,
		ArtistUuid: artistUUID,
		AlbumUuid:  albumUUID,
		Explicit:   explicit,
		TrackPath:  path,
		DurationMs: durationMs,
	})
	if err != nil {
		return nil, err
	}
	return &TrackDetails{
		TrackUUID:  res.TrackUuid,
		TrackName:  res.TrackName,
		ArtistUUID: res.ArtistUuid,
		AlbumUUID:  res.AlbumUuid,
		Explicit:   res.Explicit,
		Path:       res.TrackPath,
		DurationMs: res.DurationMs,
	}, nil
}

func (c *CatalogClient) AddTracksToAlbum(ctx context.Context, albumUUID string, tracks []AlbumTrackInput) error {
	reqTracks := make([]*catalogv1.AlbumTrackInput, 0, len(tracks))
	for _, t := range tracks {
		reqTracks = append(reqTracks, &catalogv1.AlbumTrackInput{TrackUuid: t.TrackUUID, Position: t.Position})
	}
	_, err := c.grpcWriteClient.AddTracksToAlbum(ctx, &catalogv1.AddTracksToAlbumRequest{AlbumUuid: albumUUID, Tracks: reqTracks})
	return err
}

func (c *CatalogClient) SetTrackCover(ctx context.Context, trackUUID, coverPath string) error {
	_, err := c.grpcWriteClient.SetTrackCover(ctx, &catalogv1.SetTrackCoverRequest{TrackUuid: trackUUID, CoverPath: coverPath})
	return err
}

// DeleteTrack удаляет трек из каталога и возвращает путь его файла
// в хранилище — файл удаляет gateway, у catalog_service тома нет.
func (c *CatalogClient) DeleteTrack(ctx context.Context, trackUUID string) (string, error) {
	res, err := c.grpcWriteClient.DeleteTrack(ctx, &catalogv1.DeleteTrackRequest{TrackUuid: trackUUID})
	if err != nil {
		return "", err
	}
	return res.TrackPath, nil
}

// DeleteAlbum удаляет альбом со всеми его треками и возвращает пути файлов
// хранилища (файлы треков и обложка) — удаляет их gateway.
func (c *CatalogClient) DeleteAlbum(ctx context.Context, albumUUID string) ([]string, string, error) {
	res, err := c.grpcWriteClient.DeleteAlbum(ctx, &catalogv1.DeleteAlbumRequest{AlbumUuid: albumUUID})
	if err != nil {
		return nil, "", err
	}
	return res.TrackPaths, res.CoverPath, nil
}

func (c *CatalogClient) SetAlbumCover(ctx context.Context, albumUUID, coverPath string) error {
	_, err := c.grpcWriteClient.SetAlbumCover(ctx, &catalogv1.SetAlbumCoverRequest{AlbumUuid: albumUUID, CoverPath: coverPath})
	return err
}

func (c *CatalogClient) IncrementListened(ctx context.Context, trackUUID string) error {
	_, err := c.grpcWriteClient.IncrementListened(ctx, &catalogv1.IncrementListenedRequest{TrackUuid: trackUUID})
	return err
}
