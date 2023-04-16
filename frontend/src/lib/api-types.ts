export interface User {
  id: number;
  login: string;
  name: string;
  email: string;
  roles: string[];
  thumb: string;
}

export interface Series {
  id: number;
  createdAt: number;
  lastUpdate: string;
  title: string;
  thumb: string;
}

export interface Progress {
  progress: number;
  elapsed: number;
  eta: number;
  speed: number;
}

export type TorrentStatus = 'idle' | 'download' | 'error' | 'ready'

export interface Torrent {
  id: number;
  updatedAt: string;
  name: string;
  status: TorrentStatus;
  totalLength: number;
  totalDownloadLength: number;
  bytesRead: number;
  progress: Progress;
}

export interface TorrentWithFiles extends Torrent {
  files: TorrentFile[];
}

export interface TorrentFile {
  clientIndex: number;
  selected: boolean;
  path: string;
  status: TorrentStatus;
  length: number;
}

export type ConversionStatus = 'created' | 'processing' | 'error' | 'cancelled' | 'ready'

export interface Conversion {
  id: number;
  seriesId: number;
  torrentId: number;
  torrentFileId: number;
  updatedAt: string;
  name: string;
  episodeName: string;
  command: string;
  status: ConversionStatus;
  progress: Progress;
}

export interface Episode {
  id: number;
  seriesId: number;
  conversionId: number;
  createdAt: string;
  title: string;
  episode: string;
  season: string;
  link: string;
  thumb: string;
  length: number;
  durationSec: number;
}

export interface BaseStream {
  index: number;
  name: string;
  size: number;
  lang: string;
}

export interface VideoStream extends BaseStream {
  width: number;
  height: number;
  durationSec: number;
}

export interface SubStream extends BaseStream {
  type: string;
  textLength: number;
}

export interface Analysis {
  video: VideoStream;
  audio: BaseStream[];
  sub: SubStream[];
  season: string;
  episode: string;
}

export interface ConversionPreference {
  disable?: boolean;
  stream?: number;
  file?: string;
}

export interface StartConversionFileData {
  index: number;
  episode?: string;
  season?: string;
  audio: ConversionPreference;
  sub: ConversionPreference;
}

export interface StartConversionRequest {
  torrentId: number;
  files: StartConversionFileData[];
}

export interface RoomState {
  episodeId: number | null;
  timestamp: number;
  playing: boolean;
  initiatorId?: number;
}

export interface WatcherState extends WatcherStatePartial{
  id: number;
  name: string;
  thumb: string;
}

export interface WatcherStatePartial {
  timestamp: number;
  progress: number;
  status: string;
}
