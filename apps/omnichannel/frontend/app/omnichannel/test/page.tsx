import { redirect } from 'next/navigation';

export default function OmnichannelTestRedirectPage() {
  redirect('/notifications/create');
}
