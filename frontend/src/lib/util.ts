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
