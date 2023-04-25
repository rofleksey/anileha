import axios, {AxiosProgressEvent} from 'axios';
import {AutoTorrent, SearchResult, SetSeriesQueryRequestData, StartConversionRequest, User} from 'src/lib/api-types';

const BASE_URL = import.meta.env.VITE_BASE_URL
console.log(`BASE_URL = ${BASE_URL}`)

const LONG_TIMEOUT = 30000
const SUPER_LONG_TIMEOUT = 120000
const MAX_FILE_SIZE = 5368709120;

export async function postLogin(user: string, pass: string): Promise<User> {
  const {data}: { data: User } = await axios.post(`${BASE_URL}/user/login`, {
    user,
    pass,
  }, {
    withCredentials: true,
  });
  return data;
}

export async function postLogout(): Promise<void> {
  await axios.post(`${BASE_URL}/user/logout`, {}, {
    withCredentials: true,
  });
}

export async function postModifyAccount(name: string, pass: string, email: string): Promise<User> {
  const {data}: { data: User } = await axios.post(`${BASE_URL}/user/modify`, {
    name,
    pass,
    email,
  }, {
    withCredentials: true,
  });
  return data;
}

export async function postNewUser(login: string, pass: string, email: string, roles: string[]): Promise<void> {
  await axios.post(`${BASE_URL}/owner/user`, {
    login,
    pass,
    email,
    roles,
  }, {
    withCredentials: true,
  });
}

export async function postAccountAvatar(image: File): Promise<string> {
  const formData = new FormData();
  formData.append('image', image);
  const {data}: { data: string } = await axios({
    method: 'post',
    url: `${BASE_URL}/user/avatar`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    timeout: LONG_TIMEOUT,
    maxContentLength: MAX_FILE_SIZE,
    maxBodyLength: MAX_FILE_SIZE
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
    maxContentLength: MAX_FILE_SIZE,
    maxBodyLength: MAX_FILE_SIZE
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
    url: `${BASE_URL}/admin/episodes/`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    onUploadProgress: progressCallback,
    timeout: 0,
    maxContentLength: MAX_FILE_SIZE,
    maxBodyLength: MAX_FILE_SIZE
  })
}

export async function postNewTorrentFromFile(seriesId: number, file: File, auto?: AutoTorrent): Promise<void> {
  const formData = new FormData();
  formData.append('seriesId', seriesId.toString());
  formData.append('file', file);
  if (auto) {
    formData.append('auto', JSON.stringify(auto));
  }
  await axios({
    method: 'post',
    url: `${BASE_URL}/admin/torrent/fromFile`,
    data: formData,
    headers: {'Content-Type': 'multipart/form-data'},
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
    maxContentLength: MAX_FILE_SIZE,
    maxBodyLength: MAX_FILE_SIZE
  })
}

export async function postSearchTorrents(query: string, page?: number): Promise<SearchResult[]> {
  const {data}: { data: SearchResult[] } = await axios.post(`${BASE_URL}/admin/search/torrent`, {
    query,
    page: page ?? 0,
  }, {
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
  });
  return data;
}

export async function postNewTorrentFromSearch(seriesId: number, torrentId: string, provider: string, auto?: AutoTorrent): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/fromSearch`, {
    seriesId,
    torrentId,
    provider,
    auto
  }, {
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
    maxContentLength: MAX_FILE_SIZE,
    maxBodyLength: MAX_FILE_SIZE
  });
}

export async function postStartTorrent(torrentId: number, fileIndices: number[]): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/start`, {
    id: torrentId,
    fileIndices,
  }, {
    withCredentials: true,
    timeout: LONG_TIMEOUT,
  });
}

export async function postStopTorrent(torrentId: number): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/stop`, {
    id: torrentId,
  }, {
    withCredentials: true,
  });
}

export async function postSetSeriesQuery(seriesId: number, query: SetSeriesQueryRequestData | null): Promise<void> {
  await axios.post(`${BASE_URL}/admin/search/series/setQuery`, {
    seriesId,
    query
  }, {
    withCredentials: true,
  });
}

export async function postAddTorrentsFromQuery(seriesId: number, query: SetSeriesQueryRequestData): Promise<void> {
  await axios.post(`${BASE_URL}/admin/torrent/fromQuery`, {
    seriesId,
    query
  }, {
    withCredentials: true,
    timeout: 0,
  });
}

export async function postSubText(torrentId: number, fileIndex: number, stream: number): Promise<string> {
  const {data}: { data: string } = await axios.post(`${BASE_URL}/admin/subText`, {
    id: torrentId,
    fileIndex,
    stream,
  }, {
    withCredentials: true,
    timeout: SUPER_LONG_TIMEOUT,
  });
  return data;
}

export async function postStartConversion(req: StartConversionRequest): Promise<void> {
  await axios.post(`${BASE_URL}/admin/convert/start`, req, {
    withCredentials: true,
    timeout: 0,
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
