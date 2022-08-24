import axios from "axios";
import prettyBytes from "pretty-bytes";
import {format} from "timeago.js";
import durationFormat from "format-duration";

function formatSeries(data, isAdmin) {
  return data.map((series) => {
    const details = [
      {
        id: "updated_at",
        text: format(new Date(series.updatedAt)),
      },
    ];
    if (isAdmin) {
      details.push({
        id: "torrents",
        text: "torrents",
        link: `/torrents/series/${series.id}`,
      });
      details.push({
        id: "conversions",
        text: "conversions",
        link: `/convert/series/${series.id}`,
      });
    }
    return {
      id: series.id,
      title: series.name,
      link: `/episodes/series/${series.id}`,
      details: details,
    };
  });
}

function formatTorrents(data) {
  return data.map((t) => {
    const details = [
      {
        id: "updated_at",
        text: format(new Date(t.updatedAt)),
      },
      {
        id: "length",
        text: `${prettyBytes(t.totalDowloadLength ?? 0)} / ${prettyBytes(
            t.totalLength ?? 0
        )}`,
      },
      {
        id: "status",
        text: t.status,
      },
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
          text: t.progress.speed + " fps",
        });
      }
    }
    if (t.auto) {
      details.push({
        id: "auto",
        text: "auto",
      });
    }
    return {
      id: t.id,
      title: t.name,
      link: `/torrents/${t.id}`,
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
          text: convert.progress.eta + "s remaining",
        });
      }
      if (convert.progress.speed) {
        details.push({
          id: "speed",
          text: convert.progress.speed + " fps",
        });
      }
    }
    return {
      id: convert.id,
      title: convert.name,
      link: `/convert/${convert.id}`,
      details,
    };
  });
}

function formatEpisodes(data) {
  return data.map((ep) => ({
    id: ep.id,
    title: ep.name,
    link: `/episodes/${ep.id}`,
    details: [
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
    ],
  }));
}

export async function getAllSeries(isAdmin) {
  const {data} = await axios("http://localhost:5000/series");
  return formatSeries(data, isAdmin);
}

export async function getAllTorrents() {
  const {data} = await axios("http://localhost:5000/admin/torrent");
  return formatTorrents(data);
}

export async function getAllConversions() {
  const {data} = await axios("http://localhost:5000/admin/convert");
  return formatTorrents(data);
}

export async function getTorrentsBySeriesId(seriesId) {
  const {data} = await axios(
      `http://localhost:5000/admin/torrent/series/${seriesId}`
  );
  return formatTorrents(data);
}

export async function getConversionsBySeriesId(seriesId) {
  const {data} = await axios(
      `http://localhost:5000/admin/convert/series/${seriesId}`
  );
  return formatConversions(data);
}

export async function getEpisodesBySeriesId(seriesId) {
  const {data} = await axios(
      `http://localhost:5000/series/${seriesId}/episodes`
  );
  console.log(data);
  return formatEpisodes(data);
}
