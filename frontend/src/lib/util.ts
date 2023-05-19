import {Notify} from 'quasar'

export type QuasarColumnType = {
  name: string;
  label: string;
  field: string | ((row: any) => any);
  required?: boolean;
  align?: 'left' | 'right' | 'center';
  sortable?: boolean;
  sort?: (a: any, b: any, rowA: any, rowB: any) => number;
  sortOrder?: 'ad' | 'da';
  format?: (val: any, row: any) => any;
  style?: string | ((row: any) => string);
  classes?: string | ((row: any) => string);
  headerStyle?: string;
  headerClasses?: string;
};

export interface WatchLaterItem {
  seriesId: number;
  episodeId: number;
  episodeName: string;
}

export function loadWatchLater(): WatchLaterItem[] {
  const storedStr = localStorage.getItem('watch-later');
  if (!storedStr) {
    return [];
  }
  return JSON.parse(storedStr);
}

export function saveWatchLater(seriesId: number, episodeId: number, episodeName: string) {
  const curWatchLater = loadWatchLater();
  const existingIndex = curWatchLater.findIndex((it) => it.seriesId === seriesId);
  if (existingIndex >= 0) {
    curWatchLater.splice(existingIndex, 1);
  }
  while (curWatchLater.length > 100) {
    curWatchLater.pop();
  }
  curWatchLater.unshift({
    seriesId,
    episodeId,
    episodeName,
  });
  localStorage.setItem('watch-later', JSON.stringify(curWatchLater));
}

export function showError(title: string, error: any) {
  let description: string;
  if (error.response?.data?.error) {
    description = error.response?.data?.error.toString();
  } else if (error.message) {
    description = error.message.toString();
  } else {
    description = error?.toString() ?? '';
  }
  console.log(error);
  Notify.create({
    type: 'negative',
    message: title,
    caption: description,
    timeout: 3000,
  })
}

export function showSuccess(title: string, message?: string) {
  Notify.create({
    type: 'positive',
    message: title,
    caption: message,
    timeout: 3000,
  })
}

export function showHint(title: string, message?: string) {
  Notify.create({
    type: 'warning',
    message: title,
    caption: message,
    timeout: 2000,
  })
}
