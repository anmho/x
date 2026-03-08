'use client';

import { useEffect } from 'react';
import { runDueCronJobsTick } from '@/lib/omnichannel-cron-runner';

const TICK_INTERVAL_MS = 30_000;

export function CronDaemon() {
  useEffect(() => {
    void runDueCronJobsTick();
    const timer = window.setInterval(() => {
      void runDueCronJobsTick();
    }, TICK_INTERVAL_MS);
    return () => window.clearInterval(timer);
  }, []);

  return null;
}
