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

export interface BaseStream {
  index: number;
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
  textLength: string;
}

export interface Analysis {
  video: VideoStream;
  audio: BaseStream[];
  sub: SubStream[];
}
