import {Notify} from 'quasar'

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
