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

export type FileType = 'video' | 'audio' | 'subtitle' | 'unknown'

export function getFileType(path: string): FileType {
  const lowerCase = path.toLowerCase();
  const ext = lowerCase.split('.').pop() ?? '';
  if (VIDEO_EXTENSIONS.includes(ext)) {
    return 'video';
  }
  if (AUDIO_EXTENSIONS.includes(ext)) {
    return 'audio';
  }
  if (SUBTITLE_EXTENSIONS.includes(ext)) {
    return 'subtitle';
  }
  return 'unknown';
}

const VIDEO_EXTENSIONS = [
  '3gp', '3gpp', '3g2', 'h261', 'h263', 'h264',
  'm4s', 'jpgv', 'jpm', 'jpgm', 'mj2', 'mjp2',
  'ts', 'mp4', 'mp4v', 'mpg4', 'mpeg', 'mpg',
  'mpe', 'm1v', 'm2v', 'ogv', 'qt', 'mov',
  'uvh', 'uvvh', 'uvm', 'uvvm', 'uvp', 'uvvp',
  'uvs', 'uvvs', 'uvv', 'uvvv', 'dvb', 'fvt',
  'mxu', 'm4u', 'pyv', 'uvu', 'uvvu', 'viv',
  'webm', 'f4v', 'fli', 'flv', 'm4v', 'mkv',
  'mk3d', 'mks', 'mng', 'asf', 'asx', 'vob',
  'wm', 'wmv', 'wmx', 'wvx', 'avi', 'movie',
  'smv'
];

const AUDIO_EXTENSIONS = [
  '3gpp', 'adts', 'aac', 'adp',
  'amr', 'au', 'snd', 'mid',
  'midi', 'kar', 'rmi', 'mxmf',
  'mp3', 'm4a', 'mp4a', 'mpga',
  'mp2', 'mp2a', 'm2a', 'm3a',
  'oga', 'ogg', 'spx', 'opus',
  's3m', 'sil', 'uva', 'uvva',
  'eol', 'dra', 'dts', 'dtshd',
  'lvp', 'pya', 'ecelp4800', 'ecelp7470',
  'ecelp9600', 'rip', 'wav', 'weba',
  'aif', 'aiff', 'aifc', 'caf',
  'flac', 'mka', 'm3u', 'wax',
  'wma', 'ram', 'ra', 'rmp',
  'xm'
];

const SUBTITLE_EXTENSIONS = [
  'srt', 'ssa', 'ass'
]
