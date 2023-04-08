export interface User {
  user: string;
  isAdmin: boolean;
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
  updatedAt: number;
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
  updatedAt: number;
  name: string;
  episodeName: string;
  command: string;
  status: ConversionStatus;
  progress: Progress;
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

export function isSubStream(stream: BaseStream): stream is SubStream {
  return (stream as SubStream).type !== undefined;
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
  lang?: string;
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
