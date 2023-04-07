import axios from 'axios';
import {Series, Torrent, TorrentWithFiles, User} from 'src/lib/api-types';

axios.defaults.timeout = 120000;

export const BASE_URL = import.meta.env.VITE_BASE_URL

export async function fetchMyself(): Promise<User> {
  const {data}: { data: User } = await axios.get(
    `${BASE_URL}/user/me`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchAllSeries(): Promise<Series[]> {
  const {data}: { data: Series[] } = await axios.get(
    `${BASE_URL}/series`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchAllTorrents(): Promise<Torrent[]> {
  const {data}: { data: Torrent[] } = await axios.get(
    `${BASE_URL}/admin/torrent`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchTorrentById(id: number): Promise<TorrentWithFiles> {
  const {data}: { data: TorrentWithFiles } = await axios.get(
    `${BASE_URL}/admin/torrent/${id}`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchTorrentsBySeriesId(id: number): Promise<Torrent[]> {
  const {data}: { data: Torrent[] } = await axios.get(
    `${BASE_URL}/admin/torrent/series/${id}`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchSeriesById(id: number): Promise<Series> {
  const {data}: { data: Series } = await axios.get(
    `${BASE_URL}/series/${id}`,
    {
      withCredentials: true,
    }
  );
  return data;
}
