import { LucideIcon } from 'lucide-react';
import Link from 'next/link';

export function PageIntro({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <section>
      <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
      <p className="mt-2 text-zinc-400">{description}</p>
    </section>
  );
}

export function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <article className="rounded-lg border border-zinc-800 bg-zinc-950 p-4">
      <p className="text-xs font-medium uppercase tracking-wide text-zinc-500">{title}</p>
      <p className="mt-2 text-2xl font-semibold text-zinc-100">{value}</p>
    </article>
  );
}

export function ServiceTile({
  href,
  title,
  description,
  icon: Icon,
  iconClass,
}: {
  href: string;
  title: string;
  description: string;
  icon: LucideIcon;
  iconClass: string;
}) {
  return (
    <Link
      href={href}
      className="rounded-xl border border-zinc-800 bg-zinc-950 p-4 transition hover:border-zinc-700 hover:bg-zinc-900/70"
    >
      <div className={`inline-flex h-9 w-9 items-center justify-center rounded-lg border ${iconClass}`}>
        <Icon className="h-4 w-4" />
      </div>
      <p className="mt-3 text-sm font-medium">{title}</p>
      <p className="mt-1 text-sm leading-relaxed text-zinc-400">{description}</p>
    </Link>
  );
}
