import { CronDaemon } from '@/app/omnichannel/_components/cron-daemon';

export default function OmnichannelLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <CronDaemon />
      {children}
    </>
  );
}
