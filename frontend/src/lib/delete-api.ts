import axios from 'axios';

axios.defaults.timeout = 30000;

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
