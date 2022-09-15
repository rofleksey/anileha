import axios from "axios";
import prettyBytes from "pretty-bytes";
import {format} from "timeago.js";
import durationFormat from "format-duration";
import {notify} from "@kyvg/vue3-notification";
import sanitize from "sanitize-filename";

function formatSeries(data) {
  return data.map((series) => {
    const details = [
      {
        id: "updated_at",
        text: format(new Date(series.updatedAt)),
      },
    ];
    details.push({
      id: "torrents",
      text: "torrents",
      link: `/torrents/series/${series.id}`,
      admin: true,
    });
    details.push({
      id: "conversions",
      text: "conversions",
      link: `/convert/series/${series.id}`,
      admin: true,
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
                  type: "success",
                });
              })
              .catch((err) => {
                notify({
                  title: "Failed to delete series",
                  text: err?.response?.data?.error ?? "",
                  type: "error",
                });
              });
        }
      },
      admin: true,
    });
    return {
      id: series.id,
      title: series.name,
      bg: series.thumb,
      link: `/episodes/series/${series.id}`,
      details: details,
    };
  });
}

function formatTorrents(data) {
  return data.map((t) => {
    const details = [
      {
        id: "status",
        text: t.status,
      },
      {
        id: "updated_at",
        text: format(new Date(t.updatedAt)),
      },
    ];
    if (t.status === "download") {
      details.push({
        id: "length",
        text: `${prettyBytes(t.bytesRead ?? 0)} / ${prettyBytes(
            t.totalDownloadLength ?? 0
        )}`,
      });
    } else {
      details.push({
        id: "length",
        text: `${prettyBytes(t.totalDownloadLength ?? 0)}`,
      });
    }
    if (t.status === "download") {
      if (t.progress.progress) {
        details.push({
          id: "progress",
          text: t.progress.progress + "%",
        });
      }
      if (t.progress.elapsed) {
        details.push({
          id: "elapsed",
          text: durationFormat(t.progress.elapsed * 1000) + " elapsed",
        });
      }
      if (t.progress.eta) {
        details.push({
          id: "eta",
          text: durationFormat(t.progress.eta * 1000) + " remaining",
        });
      }
      if (t.progress.speed) {
        details.push({
          id: "speed",
          text: prettyBytes(t.progress.speed) + "/s",
        });
      }
    }
    if (t.auto) {
      details.push({
        id: "auto",
        text: "auto",
      });
    }
    details.push({
      id: "files",
      text: "files",
      link: `/torrents/files/${t.id}`,
      admin: true,
    });
    if (t.status === "download") {
      details.push({
        id: "stop",
        text: "stop",
        onclick: () => {
          if (window.confirm(`Stop torrent ${t.name}?`)) {
            axios
                .post(`/admin/torrent/stop`, {
                  torrentId: t.id,
                })
                .then(() => {
                  notify({
                    title: "Stopped",
                    type: "success",
                  });
                })
                .catch((err) => {
                  notify({
                    title: "Failed to stop torrent",
                    text: err?.response?.data?.error ?? "",
                    type: "error",
                  });
                });
          }
        },
        admin: true,
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
                  type: "success",
                });
              })
              .catch((err) => {
                notify({
                  title: "Failed to delete torrent",
                  text: err?.response?.data?.error ?? "",
                  type: "error",
                });
              });
        }
      },
      admin: true,
    });
    return {
      id: t.id,
      title: t.name,
      link: `/torrents/${t.id}`,
      details,
    };
  });
}

function formatTorrentFiles(data) {
  return data.map((f) => {
    const details = [
      {
        id: "status",
        text: f.status,
      },
      {
        id: "length",
        text: prettyBytes(f.length),
      },
    ];
    if (f.season) {
      details.push({
        id: "episode",
        text: `${f.season} | ${f.episode}`,
      });
    } else {
      details.push({
        id: "episode",
        text: f.episode,
      });
    }
    return {
      id: f.id,
      title: `${f.episodeIndex}. ${f.path}`,
      link: "",
      details,
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
      {
        id: "updated_at",
        text: format(new Date(convert.updatedAt)),
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
          text: durationFormat(convert.progress.elapsed * 1000) + " elapsed",
        });
      }
      if (convert.progress.eta) {
        details.push({
          id: "eta",
          text: durationFormat(convert.progress.eta * 1000) + " remaining",
        });
      }
      if (convert.progress.speed) {
        details.push({
          id: "speed",
          text: convert.progress.speed + "x",
        });
      }
    }
    details.push({
      id: "logs",
      text: "logs",
      link: `/convert/${convert.id}/logs`,
    });
    return {
      id: convert.id,
      title: convert.name,
      link: `/convert/${convert.id}`,
      details,
    };
  });
}

function formatEpisodes(data, seriesData) {
  return data.map((ep) => {
    const details = [
      {
        id: "created_at",
        text: format(new Date(ep.createdAt)),
      },
      {
        id: "duration",
        text: durationFormat(ep.durationSec * 1000),
      },
      {
        id: "length",
        text: prettyBytes(ep.length),
      },
    ];
    details.push({
      id: "download",
      text: "download",
      onclick: () => {
        const seriesName = seriesData.name;
        const link = document.createElement("a");
        link.href = ep.link;
        link.setAttribute(
            "download",
            sanitize(`${seriesName} - ${ep.name}.mp4`)
        );
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      },
    });
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
                  type: "success",
                });
              })
              .catch((err) => {
                notify({
                  title: "Failed to delete episode",
                  text: err?.response?.data?.error ?? "",
                  type: "error",
                });
              });
        }
      },
      admin: true,
    });
    return {
      id: ep.id,
      title: ep.name,
      link: `/episodes/${ep.id}`,
      details,
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
  const {data} = await axios(`/admin/torrent/series/${seriesId}`);
  return formatTorrents(data);
}

export async function getConversionsBySeriesId(seriesId) {
  const {data} = await axios(`/admin/convert/series/${seriesId}`);
  return formatConversions(data);
}

export async function getEpisodesBySeriesId(seriesId) {
  const {data: seriesData} = await axios(`/series/${seriesId}`);
  const {data} = await axios(`/series/${seriesId}/episodes`);
  return formatEpisodes(data, seriesData);
}

export async function getTorrentFilesByTorrentId(torrentId) {
  const {data} = await axios(`/admin/torrent/${torrentId}`);
  return formatTorrentFiles(data.files);
}
