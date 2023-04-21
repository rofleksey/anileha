import axios from 'axios';
import {Conversion, Episode, Series, Torrent, TorrentWithFiles, User} from 'src/lib/api-types';

axios.defaults.timeout = 10000;

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

export async function fetchAllUsers(): Promise<User[]> {
  const {data}: { data: User[] } = await axios.get(
    `${BASE_URL}/owner/user`,
    {
      withCredentials: true,
    }
  );
  return data;
}


export async function fetchAllConversions(): Promise<Conversion[]> {
  const {data}: { data: Conversion[] } = await axios.get(
    `${BASE_URL}/admin/convert`,
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

export async function fetchConversionsBySeriesId(id: number): Promise<Conversion[]> {
  const {data}: { data: Conversion[] } = await axios.get(
    `${BASE_URL}/admin/convert/series/${id}`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchEpisodesBySeriesId(id: number): Promise<Episode[]> {
  const {data}: { data: Episode[] } = await axios.get(
    `${BASE_URL}/episodes/series/${id}`,
    {
      withCredentials: true,
    }
  );
  return data;
}

export async function fetchEpisodeById(id: number): Promise<Episode> {
  const {data}: { data: Episode } = await axios.get(
    `${BASE_URL}/episodes/${id}`,
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

export async function fetchConversionLogs(id: number): Promise<string> {
  const {data}: { data: string } = await axios.get(
    `${BASE_URL}/admin/convert/${id}/logs`,
    {
      withCredentials: true,
    }
  );
  return data;
}
