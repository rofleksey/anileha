import axios from "axios";
import prettyBytes from "pretty-bytes";
import { format } from "timeago.js";
import durationFormat from "format-duration";
import { notify } from "@kyvg/vue3-notification";

function formatSeries(data) {
  return data.map((series) => {
    const details = [
      {
        id: "updated_at",
        text: format(new Date(series.updatedAt))
      }
    ];
    details.push({
      id: "torrents",
      text: "torrents",
      link: `/torrents/series/${series.id}`,
      admin: true
    });
    details.push({
      id: "conversions",
      text: "conversions",
      link: `/convert/series/${series.id}`,
      admin: true
    });
    details.push({
      id: "delete",
      text: "delete",
      onclick: () => {
        if (window.confirm(`Delete series ${series.name}?`)) {
          axios
            .delete(`/admin/series/${series.id}`)
            .then(() => {
              notify({
                title: "Deleted",
                type: "success"
              });
            })
            .catch((err) => {
              notify({
                title: "Failed to delete series",
                text: err?.response?.data?.error ?? "",
                type: "error"
              });
            });
        }
      },
      admin: true
    });
    return {
      id: series.id,
      title: series.name,
      link: `/episodes/series/${series.id}`,
      details: details
    };
  });
}

function formatTorrents(data) {
  return data.map((t) => {
    const details = [
      {
        id: "status",
        text: t.status
      },
      {
        id: "updated_at",
        text: format(new Date(t.updatedAt))
      },
      {
        id: "length",
        text: `${prettyBytes(t.totalDowloadLength ?? 0)} / ${prettyBytes(
          t.totalLength ?? 0
        )}`
      }
    ];
    if (t.status === "processing") {
      if (t.progress.progress) {
        details.push({
          id: "progress",
          text: t.progress.progress + "%",
        });
      }
      if (t.progress.elapsed) {
        details.push({
          id: "elapsed",
          text: t.progress.elapsed + "s elapsed",
        });
      }
      if (t.progress.eta) {
        details.push({
          id: "eta",
          text: t.progress.eta + "s remaining",
        });
      }
      if (t.progress.speed) {
        details.push({
          id: "speed",
          text: t.progress.speed + " fps"
        });
      }
    }
    if (t.auto) {
      details.push({
        id: "auto",
        text: "auto"
      });
    }
    if (t.status === "processing") {
      details.push({
        id: "stop",
        text: "stop",
        onclick: () => {
          if (window.confirm(`Stop torrent ${t.name}?`)) {
            axios
              .post(`/admin/torrent/stop`, {
                torrentId: t.id
              })
              .then(() => {
                notify({
                  title: "Stopped",
                  type: "success"
                });
              })
              .catch((err) => {
                notify({
                  title: "Failed to stop torrent",
                  text: err?.response?.data?.error ?? "",
                  type: "error"
                });
              });
          }
        },
        admin: true
      });
    }
    details.push({
      id: "delete",
      text: "delete",
      onclick: () => {
        if (window.confirm(`Delete torrent ${t.name}?`)) {
          axios
            .delete(`/admin/torrent/${t.id}`)
            .then(() => {
              notify({
                title: "Deleted",
                type: "success"
              });
            })
            .catch((err) => {
              notify({
                title: "Failed to delete torrent",
                text: err?.response?.data?.error ?? "",
                type: "error"
              });
            });
        }
      },
      admin: true
    });
    return {
      id: t.id,
      title: t.name,
      link: `/torrents/${t.id}`,
      details
    };
  });
}

function formatConversions(data) {
  return data.map((convert) => {
    const details = [
      {
        id: "status",
        text: convert.status,
      },
    ];
    if (convert.status === "processing") {
      if (convert.progress.progress) {
        details.push({
          id: "progress",
          text: convert.progress.progress + "%",
        });
      }
      if (convert.progress.elapsed) {
        details.push({
          id: "elapsed",
          text: convert.progress.elapsed + "s elapsed",
        });
      }
      if (convert.progress.eta) {
        details.push({
          id: "eta",
          text: convert.progress.eta + "s remaining"
        });
      }
      if (convert.progress.speed) {
        details.push({
          id: "speed",
          text: convert.progress.speed + " fps"
        });
      }
    }
    details.push({
      id: "logs",
      text: "logs",
      link: `/convert/${convert.id}/logs`
    });
    return {
      id: convert.id,
      title: convert.name,
      link: `/convert/${convert.id}`,
      details
    };
  });
}

function formatEpisodes(data) {
  return data.map((ep) => {
    const details = [
      {
        id: "created_at",
        text: format(new Date(ep.createdAt))
      },
      {
        id: "duration",
        text: durationFormat(ep.durationSec * 1000)
      },
      {
        id: "length",
        text: prettyBytes(ep.length)
      }
    ];
    details.push({
      id: "delete",
      text: "delete",
      onclick: () => {
        if (window.confirm(`Delete episode ${ep.name}?`)) {
          axios
            .delete(`/admin/episodes/${ep.id}`)
            .then(() => {
              notify({
                title: "Deleted",
                type: "success"
              });
            })
            .catch((err) => {
              notify({
                title: "Failed to delete episode",
                text: err?.response?.data?.error ?? "",
                type: "error"
              });
            });
        }
      },
      admin: true
    });
    return {
      id: ep.id,
      title: ep.name,
      link: `/episodes/${ep.id}`,
      details
    };
  });
}

export async function getAllSeries() {
  const { data } = await axios("/series");
  return formatSeries(data);
}

export async function getAllTorrents() {
  const { data } = await axios("/admin/torrent");
  return formatTorrents(data);
}

export async function getAllConversions() {
  const { data } = await axios("/admin/convert");
  return formatConversions(data);
}

export async function getTorrentsBySeriesId(seriesId) {
  const { data } = await axios(
    `/admin/torrent/series/${seriesId}`
  );
  return formatTorrents(data);
}

export async function getConversionsBySeriesId(seriesId) {
  const { data } = await axios(
    `/admin/convert/series/${seriesId}`
  );
  return formatConversions(data);
}

export async function getEpisodesBySeriesId(seriesId) {
  const { data } = await axios(
    `/series/${seriesId}/episodes`
  );
  return formatEpisodes(data);
}
