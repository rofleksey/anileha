import axios, {AxiosProgressEvent} from 'axios';
import {Analysis, StartConversionRequest, User} from 'src/lib/api-types';

const BASE_URL = import.meta.env.VITE_BASE_URL
console.log(`BASE_URL = ${BASE_URL}`)

const LONG_TIMEOUT = 30000
const SUPER_LONG_TIMEOUT = 120000

export async function postLogin(user: string, pass: string): Promise<User> {
  const {data}: { data: User } = await axios.post(`${BASE_URL}/user/login`, {
    user,
    pass,
  }, {
    withCredentials: true,
  });
  return data;
}

export async function postNewSeries(title: string, thumb: File): Promise<void> {
  const formData = new FormData();
  formData.append('title', title);
  formData.append('thumb', thumb);
  await axios({
    method: 'post',
    url: `${BASE_URL}/admin/series`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    timeout: LONG_TIMEOUT,
    maxContentLength: Infinity,
    maxBodyLength: Infinity
  })
}

export async function postNewEpisode(seriesId: number | null, file: File, title: string, season: string | null,
                                     episode: string | null, progressCallback: (e: AxiosProgressEvent) => void): Promise<void> {
  const formData = new FormData();
  formData.append('title', title);
  formData.append('file', file);
  if (seriesId) {
    formData.append('seriesId', seriesId.toString())
  }
  if (season) {
    formData.append('season', season.toString())
  }
  if (episode) {
    formData.append('episode', episode.toString())
  }
  await axios({
    method: 'post',
    url: `${BASE_URL}/admin/episodes`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    onUploadProgress: progressCallback,
    timeout: 0,
    maxContentLength: Infinity,
    maxBodyLength: Infinity
  })
}

export async function postNewTorrent(seriesId: number, file: File): Promise<void> {
  const formData = new FormData();
  formData.append('seriesId', seriesId.toString());
  formData.append('file', file);
  await axios({
    method: 'post',
    url: `${BASE_URL}/admin/torrent`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
    maxContentLength: Infinity,
    maxBodyLength: Infinity
  })
}

export async function postStartTorrent(torrentId: number, fileIndices: number[]): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/start`, {
    id: torrentId,
    fileIndices,
  }, {
    withCredentials: true,
  });
}

export async function postStopTorrent(torrentId: number): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/stop`, {
    id: torrentId,
  }, {
    withCredentials: true,
  });
}

export async function postAnalyze(torrentId: number, fileIndex: number): Promise<Analysis> {
  const {data}: { data: Analysis } = await axios.post(`${BASE_URL}/admin/analyze`, {
    id: torrentId,
    fileIndex,
  }, {
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
  });
  return data;
}

export async function postStartConversion(req: StartConversionRequest): Promise<void> {
  await axios.post(`${BASE_URL}/admin/convert/start`, req, {
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
  });
}

export async function refreshEpisodeThumb(id: number): Promise<void> {
  await axios.post(`${BASE_URL}/admin/episodes/refreshThumb`, {
    id
  }, {
    withCredentials: true,
    timeout: LONG_TIMEOUT,
  });
}
