export type Period = 'short_term' | 'medium_term' | 'long_term'

export interface ArtistCount {
  artist_id: string
  artist_name: string
  play_count: number
}

export interface TrackCount {
  track_id: string
  track_name: string
  artist_name: string
  play_count: number
}

export interface TopArtistsResponse {
  period: Period
  snapshot_date?: string
  artists: ArtistCount[]
}

export interface GenreCount {
  genre: string
  count: number
}

export interface GenresResponse {
  period: Period
  genres: GenreCount[]
}

export interface HeatmapCell {
  day_of_week: number
  hour: number
  count: number
}

export interface HeatmapResponse {
  cells: HeatmapCell[]
}

export interface StorySlide {
  type: string
  title: string
  value?: string
  caption?: string
}

export interface StoryResponse {
  slides: StorySlide[]
}
