import axios from 'axios';
import {Analysis, User} from 'src/lib/api-types';

axios.defaults.timeout = 30000;

const BASE_URL = import.meta.env.VITE_BASE_URL
console.log(`BASE_URL = ${BASE_URL}`)

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
  });
  return data;
}
