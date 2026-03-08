'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import {
  BellRing,
  Bot,
  Calendar,
  Check,
  ChevronRight,
  ChevronsLeft,
  Cloud,
  FileText,
  FolderKanban,
  Globe,
  Key,
  LogOut,
  Megaphone,
  Network,
  PanelLeft,
  Plus,
  Rocket,
  Search,
  Send,
  Settings,
  Shield,
  Smartphone,
  Workflow,
  X,
} from 'lucide-react';
import { CATALOG_PROJECTS } from '@/lib/project-catalog';

export type NavSection =
  | 'home'
  | 'omnichannel'
  | 'omnichannelTest'
  | 'omnichannelStatus'
  | 'recipients'
  | 'cronJobs'
  | 'campaigns'
  | 'emulator'
  | 'applications'
  | 'secrets'
  | 'agents'
  | 'settings'
  | 'deployments'
  | 'projects'
  | 'templates'
  | 'notifications'
  | 'workflows'
  | 'domains';

// ─── nav structure ────────────────────────────────────────────────────────────

const OMNICHANNEL_ITEMS = [
  { href: '/omnichannel', label: 'Overview', icon: FolderKanban, section: 'omnichannel' as NavSection },
  { href: '/omnichannel/test', label: 'Test', icon: BellRing, section: 'omnichannelTest' as NavSection },
  { href: '/notifications/create', label: 'Send Notification', icon: Send, section: 'notifications' as NavSection },
  { href: '/templates', label: 'Templates', icon: FileText, section: 'templates' as NavSection },
  { href: '/deployments', label: 'Status', icon: Rocket, section: 'omnichannelStatus' as NavSection },
  { href: '/omnichannel/recipients', label: 'Recipients', icon: Network, section: 'recipients' as NavSection },
  { href: '/omnichannel/cron-jobs', label: 'Cron Jobs', icon: Calendar, section: 'cronJobs' as NavSection },
  { href: '/omnichannel/campaigns', label: 'Campaigns', icon: Megaphone, section: 'campaigns' as NavSection },
  { href: '/omnichannel/emulator', label: 'App Emulator', icon: Smartphone, section: 'emulator' as NavSection },
  { href: '/workflows', label: 'Workflows', icon: Workflow, section: 'workflows' as NavSection },
];

const MAIN_ITEMS = [
  { href: '/applications', label: 'Applications', section: 'applications' as NavSection, icon: Network },
  { href: '/secrets', label: 'Secrets', section: 'secrets' as NavSection, icon: Shield },
  { href: '/agents', label: 'Agents', section: 'agents' as NavSection, icon: Bot },
  { href: '/domains', label: 'Domains', section: 'domains' as NavSection, icon: Globe },
];

const CMD_ITEMS = [
  { href: '/', label: 'Overview', section: 'Pages' },
  { href: '/omnichannel', label: 'Omnichannel', section: 'Pages' },
  { href: '/omnichannel/test', label: 'Omnichannel Test', section: 'Pages' },
  { href: '/omnichannel/recipients', label: 'Recipients', section: 'Pages' },
  { href: '/omnichannel/cron-jobs', label: 'Cron Jobs', section: 'Pages' },
  { href: '/omnichannel/campaigns', label: 'Campaigns', section: 'Pages' },
  { href: '/omnichannel/emulator', label: 'App Emulator', section: 'Pages' },
  { href: '/deployments', label: 'Deployments', section: 'Pages' },
  { href: '/applications', label: 'Applications', section: 'Pages' },
  { href: '/secrets', label: 'Secrets', section: 'Pages' },
  { href: '/templates', label: 'Templates', section: 'Pages' },
  { href: '/agents', label: 'Agents', section: 'Pages' },
  { href: '/workflows', label: 'Workflows', section: 'Pages' },
  { href: '/domains', label: 'Domains', section: 'Pages' },
  { href: '/settings', label: 'Settings', section: 'Pages' },
  { href: '/projects', label: 'Projects', section: 'Pages' },
  { href: '/notifications/create', label: 'Send Notification', section: 'Actions' },
  { href: '/settings?tab=integrations', label: 'Integrations', section: 'Settings' },
];

const PAGE_TITLES: Partial<Record<NavSection, { title: string; description: string }>> = {
  home: { title: 'Overview', description: 'Platform dashboard and quick access.' },
  omnichannel: { title: 'Omnichannel', description: 'Operate messaging across testing, templates, recipients, schedules, and campaigns.' },
  omnichannelTest: { title: 'Omnichannel Test', description: 'Run test sends across channels without editing templates.' },
  omnichannelStatus: { title: 'Status', description: 'Runtime delivery status, queue health, and workflow state.' },
  recipients: { title: 'Recipients', description: 'Manage target audiences and recipient groups.' },
  cronJobs: { title: 'Cron Jobs', description: 'Configure recurring notification schedules.' },
  campaigns: { title: 'Campaigns', description: 'Plan and track campaign-level sends and rollout status.' },
  emulator: { title: 'App Emulator', description: 'Inspect app channel notifications delivered to the local relay.' },
  applications: { title: 'Applications', description: 'All deployed apps across environments.' },
  deployments: { title: 'Deployments', description: 'Service deployments and Temporal workflows.' },
  secrets: { title: 'Secrets', description: 'Account-level secrets shared across services.' },
  agents: { title: 'Agents', description: 'Autonomous agents running on your behalf.' },
  workflows: { title: 'Workflows', description: 'Temporal workflow orchestration.' },
  templates: { title: 'Templates', description: 'Reusable notification content blocks.' },
  notifications: { title: 'Send Notification', description: 'Compose and dispatch notifications.' },
  domains: { title: 'Domains', description: 'DNS management across all providers.' },
  settings: { title: 'Settings', description: 'Account, integrations, and security.' },
  projects: { title: 'Projects', description: 'Project health, observability, and billing.' },
};

// ─── project color ────────────────────────────────────────────────────────────

const PROJECT_COLORS = ['bg-blue-500', 'bg-violet-500', 'bg-emerald-500', 'bg-orange-500', 'bg-pink-500', 'bg-cyan-500'];
function projectColor(id: string) {
  return PROJECT_COLORS[id.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % PROJECT_COLORS.length];
}

// ─── AppShell ─────────────────────────────────────────────────────────────────

export function AppShell({ active, children }: { active?: NavSection; children?: React.ReactNode }) {
  const [projectOpen, setProjectOpen] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const [cmdOpen, setCmdOpen] = useState(false);
  const [omnichannelOpen, setOmnichannelOpen] = useState(true);
  const [projectSearch, setProjectSearch] = useState('');
  const [cmdSearch, setCmdSearch] = useState('');
  const [selectedProjectId, setSelectedProjectId] = useState(CATALOG_PROJECTS[0]?.id ?? '');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const projectBtnRef = useRef<HTMLButtonElement>(null);
  const projectMenuRef = useRef<HTMLDivElement>(null);
  const accountRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const stored = localStorage.getItem('console:selected-project');
    if (stored && CATALOG_PROJECTS.some((p) => p.id === stored)) setSelectedProjectId(stored);

    const collapsedStored = localStorage.getItem('sidebar:collapsed');
    if (collapsedStored !== null) setSidebarCollapsed(collapsedStored === 'true');
  }, []);

  useEffect(() => {
    function onMouseDown(e: MouseEvent) {
      const t = e.target as Node;
      if (projectMenuRef.current && !projectMenuRef.current.contains(t) && projectBtnRef.current && !projectBtnRef.current.contains(t)) {
        setProjectOpen(false); setProjectSearch('');
      }
      if (accountRef.current && !accountRef.current.contains(t)) setAccountOpen(false);
    }
    document.addEventListener('mousedown', onMouseDown);
    return () => document.removeEventListener('mousedown', onMouseDown);
  }, []);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') { e.preventDefault(); setCmdOpen((v) => !v); setCmdSearch(''); }
      if (e.key === 'Escape') { setCmdOpen(false); setProjectOpen(false); setAccountOpen(false); }
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  function toggleSidebar() {
    setSidebarCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem('sidebar:collapsed', String(next));
      return next;
    });
  }

  const currentProject = CATALOG_PROJECTS.find((p) => p.id === selectedProjectId);
  const filteredProjects = CATALOG_PROJECTS.filter(
    (p) => p.label.toLowerCase().includes(projectSearch.toLowerCase()) || p.id.toLowerCase().includes(projectSearch.toLowerCase()),
  );

  function selectProject(id: string) {
    setSelectedProjectId(id);
    localStorage.setItem('console:selected-project', id);
    setProjectOpen(false); setProjectSearch('');
  }

  const filteredCmdItems = CMD_ITEMS.filter(
    (i) => !cmdSearch || i.label.toLowerCase().includes(cmdSearch.toLowerCase()) || i.section.toLowerCase().includes(cmdSearch.toLowerCase()),
  );

  const pageInfo = active ? PAGE_TITLES[active] : undefined;

  function sidebarLink(href: string, label: string, Icon: React.ComponentType<{ className?: string }>, section: NavSection) {
    const isActive = active === section;
    if (sidebarCollapsed) {
      return (
        <Link key={href} href={href} title={label} className={`flex items-center justify-center rounded-lg py-2 transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`}>
          <Icon className="h-4 w-4 shrink-0" />
        </Link>
      );
    }
    return (
      <Link key={href} href={href} className={`flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`}>
        <Icon className="h-4 w-4 shrink-0" />
        {label}
      </Link>
    );
  }

  const sidebarWidth = sidebarCollapsed ? 'w-14' : 'w-56';
  const topbarLeftWidth = 'w-56';

  return (
    <div className="flex h-screen flex-col overflow-hidden app-shell">

      {/* ── Top bar ────────────────────────────────────────────────── */}
      <header className="flex h-14 shrink-0 items-center gap-4 border-b border-zinc-800 px-4">

        {/* Left: logo + project */}
        <div className={`flex ${topbarLeftWidth} shrink-0 items-center gap-2 transition-all duration-200`}>
          <Link href="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-black border border-zinc-700">
              <svg viewBox="0 0 24 24" className="h-3.5 w-3.5 fill-white" xmlns="http://www.w3.org/2000/svg">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-4.714-6.231-5.401 6.231H2.744l7.73-8.835L1.254 2.25H8.08l4.261 5.638 5.903-5.638Zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
              </svg>
            </div>
            <span className="text-sm font-medium text-zinc-500">Platform</span>
          </Link>
          <span className="text-zinc-700 select-none">/</span>

          {/* Project selector */}
          <div className="relative">
            <button ref={projectBtnRef} onClick={() => { setProjectOpen((v) => !v); setProjectSearch(''); }}
              className="inline-flex items-center gap-1.5 rounded-md px-1.5 py-1 text-sm text-zinc-200 hover:bg-zinc-900 transition-colors">
              <div className={`h-5 w-5 rounded-full ${projectColor(selectedProjectId)} flex items-center justify-center text-[10px] font-bold text-white shrink-0`}>
                {currentProject?.label?.[0] ?? 'P'}
              </div>
              <span className="max-w-[80px] truncate font-medium">{currentProject?.label ?? 'Select'}</span>
              <svg className="h-4 w-4 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
              </svg>
            </button>

            <div ref={projectMenuRef} className={`absolute left-0 top-full mt-1 w-72 overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl transition-all duration-150 origin-top-left z-50 ${projectOpen ? 'opacity-100 scale-100 pointer-events-auto' : 'opacity-0 scale-95 pointer-events-none'}`}>
              <div className="border-b border-zinc-800 px-3 py-2.5">
                <div className="flex items-center gap-2">
                  <Search className="h-3.5 w-3.5 text-zinc-500 shrink-0" />
                  <input type="text" placeholder="Find project..." value={projectSearch} onChange={(e) => setProjectSearch(e.target.value)}
                    className="flex-1 bg-transparent text-sm text-zinc-200 placeholder-zinc-500 outline-none" />
                  <kbd className="text-[10px] text-zinc-600 bg-zinc-800 px-1 py-0.5 rounded font-mono">Esc</kbd>
                </div>
              </div>
              <div className="py-1 max-h-56 overflow-y-auto">
                {filteredProjects.length === 0 ? (
                  <p className="px-3 py-4 text-center text-xs text-zinc-500">No projects found</p>
                ) : filteredProjects.map((project) => (
                  <button key={project.id} onClick={() => selectProject(project.id)} className="w-full flex items-center gap-3 px-3 py-2.5 text-left hover:bg-zinc-900 transition-colors">
                    <div className={`h-6 w-6 rounded-full ${projectColor(project.id)} flex items-center justify-center text-[10px] font-bold text-white shrink-0`}>{project.label[0]}</div>
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-zinc-100 truncate">{project.label}</p>
                      <p className="text-xs text-zinc-500 font-mono truncate">{project.id}</p>
                    </div>
                    {selectedProjectId === project.id && <Check className="h-3.5 w-3.5 text-blue-400 shrink-0" />}
                  </button>
                ))}
              </div>
              <div className="border-t border-zinc-800 p-1.5">
                <Link href="/projects" onClick={() => setProjectOpen(false)} className="flex items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition-colors">
                  <Plus className="h-3.5 w-3.5" />Create project
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Center: search */}
        <div className="flex flex-1 justify-center">
          <button onClick={() => { setCmdOpen(true); setCmdSearch(''); }}
            className="flex w-full max-w-md items-center gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 px-4 py-2 text-sm text-zinc-500 hover:border-zinc-700 hover:bg-zinc-900 transition-colors group">
            <Search className="h-4 w-4 shrink-0 group-hover:text-zinc-400" />
            <span className="flex-1 text-left">
              {pageInfo ? (
                <span><span className="text-zinc-600">{pageInfo.title} · </span>Search…</span>
              ) : 'Search pages, actions…'}
            </span>
            <span className="flex items-center gap-0.5 rounded-md bg-zinc-800 px-1.5 py-0.5 text-[11px] text-zinc-600 font-mono">⌘K</span>
          </button>
        </div>

        {/* Right: account */}
        <div className="flex w-56 items-center justify-end">
          <div ref={accountRef} className="relative">
            <button onClick={() => setAccountOpen((v) => !v)}
              className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-xs font-semibold text-white hover:bg-blue-500 transition-colors ring-2 ring-transparent hover:ring-blue-500/30">
              AH
            </button>
            <div className={`absolute right-0 top-full mt-2 w-64 overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl transition-all duration-150 origin-top-right z-50 ${accountOpen ? 'opacity-100 scale-100 pointer-events-auto' : 'opacity-0 scale-95 pointer-events-none'}`}>
              <div className="flex items-center gap-3 border-b border-zinc-800 px-4 py-3.5">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-600 text-sm font-semibold text-white">AH</div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-zinc-100">Andrew Ho</p>
                  <p className="truncate text-xs text-zinc-500">andrew@example.com</p>
                </div>
              </div>
              <div className="p-1.5">
                <Link href="/settings" onClick={() => setAccountOpen(false)} className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors">
                  <Settings className="h-4 w-4 text-zinc-500" />Account Settings
                </Link>
                <Link href="/settings?tab=integrations" onClick={() => setAccountOpen(false)} className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors">
                  <Cloud className="h-4 w-4 text-zinc-500" />Integrations
                  <span className="ml-auto rounded-full bg-zinc-800 px-1.5 py-0.5 text-[10px] text-zinc-500">Manage</span>
                </Link>
                <Link href="/secrets" onClick={() => setAccountOpen(false)} className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors">
                  <Key className="h-4 w-4 text-zinc-500" />Secrets
                </Link>
              </div>
              <div className="border-t border-zinc-800 p-1.5">
                <Link href="/api/auth/logout" className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-zinc-400 hover:bg-zinc-900 hover:text-rose-400 transition-colors">
                  <LogOut className="h-4 w-4" />Sign out
                </Link>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* ── Body ────────────────────────────────────────────────────── */}
      <div className="flex flex-1 overflow-hidden">

        {/* Sidebar */}
        <aside className={`flex ${sidebarWidth} shrink-0 flex-col border-r border-zinc-800 app-shell-bg transition-all duration-200`}>

          {/* Toggle button */}
          <div className={`flex shrink-0 items-center border-b border-zinc-800/60 px-2 py-2 ${sidebarCollapsed ? 'justify-center' : 'justify-start'}`}>
            <button
              onClick={toggleSidebar}
              title={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              className="flex items-center justify-center rounded-md p-1.5 text-zinc-500 hover:bg-zinc-900 hover:text-zinc-300 transition-colors"
            >
              {sidebarCollapsed ? <PanelLeft className="h-4 w-4" /> : <ChevronsLeft className="h-4 w-4" />}
            </button>
          </div>

          <nav className={`flex-1 overflow-y-auto space-y-0.5 px-2 py-4 ${sidebarCollapsed ? 'px-1' : 'px-2'}`}>

            {/* Omnichannel */}
            <div className="mb-1">
              {!sidebarCollapsed && (
                <button onClick={() => setOmnichannelOpen((v) => !v)}
                  className="flex w-full items-center gap-2 rounded-lg px-3 py-1.5 text-[11px] font-semibold uppercase tracking-widest text-zinc-500 hover:text-zinc-300 transition-colors">
                  <ChevronRight className={`h-3 w-3 transition-transform duration-150 ${omnichannelOpen ? 'rotate-90' : ''}`} />
                  Omnichannel
                </button>
              )}
              {(sidebarCollapsed || omnichannelOpen) && (
                <div className={`mt-0.5 space-y-0.5 ${sidebarCollapsed ? '' : 'pl-1.5'}`}>
                  {OMNICHANNEL_ITEMS.map((item) => {
                    const isActive = active === item.section;
                    if (sidebarCollapsed) {
                      return (
                        <Link key={item.href} href={item.href} title={item.label}
                          className={`flex items-center justify-center rounded-lg py-2 transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`}>
                          <item.icon className="h-4 w-4 shrink-0" />
                        </Link>
                      );
                    }
                    return (
                      <Link key={item.href} href={item.href}
                        className={`flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`}>
                        <item.icon className="h-4 w-4 shrink-0" />
                        {item.label}
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>

            <div className="my-2 border-t border-zinc-800/70" />

            {MAIN_ITEMS.map((item) => sidebarLink(item.href, item.label, item.icon, item.section))}

            <div className="my-2 border-t border-zinc-800/70" />

            {sidebarLink('/settings', 'Settings', Settings, 'settings')}
          </nav>

          {/* Account at bottom */}
          <div className="border-t border-zinc-800 px-2 py-3">
            {sidebarCollapsed ? (
              <div className="flex items-center justify-center rounded-lg py-2 hover:bg-zinc-900 transition-colors cursor-default">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-blue-600 text-xs font-semibold text-white">AH</div>
              </div>
            ) : (
              <div className="flex items-center gap-2.5 rounded-lg px-3 py-2 hover:bg-zinc-900 transition-colors cursor-default">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-blue-600 text-xs font-semibold text-white">AH</div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-zinc-200">Andrew Ho</p>
                  <p className="truncate text-xs text-zinc-500">andrew@example.com</p>
                </div>
              </div>
            )}
          </div>
        </aside>

        {/* Main */}
        <main className="flex flex-1 flex-col overflow-hidden">
          {/* Page title bar */}
          {pageInfo && (
            <div className="flex shrink-0 items-center justify-between border-b border-zinc-800/60 px-8 py-3">
              <div>
                <p className="text-sm font-semibold text-zinc-100">{pageInfo.title}</p>
                <p className="text-xs text-zinc-500">{pageInfo.description}</p>
              </div>
            </div>
          )}
          <div className="flex-1 overflow-y-auto">
            {children}
          </div>
        </main>
      </div>

      {/* ── Command palette ────────────────────────────────────────── */}
      {cmdOpen && (
        <div className="fixed inset-0 z-50">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setCmdOpen(false)} />
          <div className="relative mx-auto mt-[15vh] max-w-lg px-4">
            <div className="overflow-hidden rounded-xl border border-zinc-700 bg-zinc-950 shadow-2xl">
              <div className="flex items-center gap-3 border-b border-zinc-800 px-4 py-3">
                <Search className="h-4 w-4 text-zinc-400 shrink-0" />
                <input autoFocus type="text" placeholder="Search pages, projects, actions..." value={cmdSearch} onChange={(e) => setCmdSearch(e.target.value)}
                  className="flex-1 bg-transparent text-sm text-zinc-100 placeholder-zinc-500 outline-none" />
                <button onClick={() => setCmdOpen(false)} className="text-zinc-500 hover:text-zinc-300 transition-colors">
                  <X className="h-4 w-4" />
                </button>
              </div>
              <div className="max-h-72 overflow-y-auto p-1.5">
                {filteredCmdItems.length === 0 ? (
                  <p className="px-3 py-6 text-center text-sm text-zinc-500">No results</p>
                ) : filteredCmdItems.map((item) => (
                  <Link key={item.href} href={item.href} onClick={() => setCmdOpen(false)}
                    className="flex items-center justify-between rounded-lg px-3 py-2.5 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors">
                    <span>{item.label}</span>
                    <span className="text-xs text-zinc-600">{item.section}</span>
                  </Link>
                ))}
              </div>
              <div className="border-t border-zinc-800 px-4 py-2 flex items-center gap-4 text-[11px] text-zinc-600">
                <span><kbd className="bg-zinc-800 rounded px-1 text-zinc-500">↵</kbd> Open</span>
                <span><kbd className="bg-zinc-800 rounded px-1 text-zinc-500">Esc</kbd> Close</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export { AppShell as AppNav };
