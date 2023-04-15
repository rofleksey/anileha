import axios from 'axios';

const BASE_URL = import.meta.env.VITE_BASE_URL
console.log(`BASE_URL = ${BASE_URL}`)

export async function deleteSeries(id: number): Promise<void> {
  await axios({
    method: 'delete',
    url: `${BASE_URL}/admin/series/${id}`,
    withCredentials: true,
  })
}

export async function deleteTorrent(id: number): Promise<void> {
  await axios({
    method: 'delete',
    url: `${BASE_URL}/admin/torrent/${id}`,
    withCredentials: true,
  })
}

export async function deleteEpisode(id: number): Promise<void> {
  await axios({
    method: 'delete',
    url: `${BASE_URL}/admin/episodes/${id}`,
    withCredentials: true,
  })
}
