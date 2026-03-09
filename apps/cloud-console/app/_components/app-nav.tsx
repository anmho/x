'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Command } from 'cmdk';
import {
  Bot,
  Calendar,
  Check,
  ChevronDown,
  ChevronRight,
  ChevronsLeft,
  Cloud,
  Database,
  FolderKanban,
  Globe,
  Key,
  LogOut,
  Monitor,
  Moon,
  Network,
  PanelLeft,
  Plus,
  Rocket,
  Search,
  Send,
  Settings,
  Shield,
  Sparkles,
  Sun,
  Wrench,
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
  | 'campaigns'
  | 'cronJobs'
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
  { href: '/notifications/create', label: 'Schedule Notification', icon: Send, section: 'notifications' as NavSection },
  { href: '/api-keys', label: 'Manage API Keys', icon: Key, section: 'secrets' as NavSection },
  { href: '/deployments', label: 'Notification Status', icon: Rocket, section: 'omnichannelStatus' as NavSection },
];

const TEMPORAL_UI_URL = (typeof process !== 'undefined' && process.env.NEXT_PUBLIC_TEMPORAL_UI_URL) || 'http://localhost:8233';

const MAIN_ITEMS = [
  { href: '/cron-jobs', label: 'Cron Jobs', section: 'cronJobs' as NavSection, icon: Calendar, external: false },
  { href: TEMPORAL_UI_URL, label: 'Workflows', section: 'workflows' as NavSection, icon: Workflow, external: true },
  { href: '/applications', label: 'Applications', section: 'applications' as NavSection, icon: Network, external: false },
  { href: '/secrets', label: 'Secrets', section: 'secrets' as NavSection, icon: Shield, external: false },
  { href: '/agents', label: 'Agents', section: 'agents' as NavSection, icon: Bot, external: false },
  { href: '/domains', label: 'Domains', section: 'domains' as NavSection, icon: Globe, external: false },
];

const CMD_ITEMS = [
  { href: '/', label: 'Overview', section: 'Pages' },
  { href: '/omnichannel', label: 'Omnichannel', section: 'Pages' },
  { href: '/notifications/create', label: 'Schedule Notification', section: 'Pages' },
  { href: '/api-keys', label: 'Manage API Keys', section: 'Pages' },
  { href: '/deployments', label: 'Notification Status', section: 'Pages' },
  { href: '/cron-jobs', label: 'Cron Jobs', section: 'Pages' },
  { href: '/applications', label: 'Applications', section: 'Pages' },
  { href: '/secrets', label: 'Secrets', section: 'Pages' },
  { href: '/agents', label: 'Agents', section: 'Pages' },
  { href: TEMPORAL_UI_URL, label: 'Workflows (Temporal)', section: 'Pages' },
  { href: '/domains', label: 'Domains', section: 'Pages' },
  { href: '/settings', label: 'Settings', section: 'Pages' },
  { href: '/projects', label: 'Projects', section: 'Pages' },
  { href: '/settings?tab=integrations', label: 'Integrations', section: 'Settings' },
];

const CMD_GROUPS = CMD_ITEMS.reduce((groups, item) => {
  const existingGroup = groups.find((group) => group.section === item.section);
  if (existingGroup) {
    existingGroup.items.push(item);
    return groups;
  }
  groups.push({ section: item.section, items: [item] });
  return groups;
}, [] as Array<{ section: string; items: typeof CMD_ITEMS }>);

const CMD_ITEM_ORDER = new Map(CMD_ITEMS.map((item, index) => [item.href, index]));

const PROJECT_SHORTCUTS = [
  { id: 'build-machines', label: 'Build Machines', description: 'Build and Deployment settings', href: '/deployments', icon: Wrench },
  { id: 'storage', label: 'Storage', description: 'Secrets and stateful resources', href: '/secrets', icon: Database },
  { id: 'queued-prod', label: 'Queued Production Deployments', description: 'Deployment queue and rollout status', href: '/deployments', icon: Sparkles },
];

const PAGE_TITLES: Partial<Record<NavSection, { title: string; description: string }>> = {
  home: { title: 'Overview', description: 'Platform dashboard and quick access.' },
  omnichannel: { title: 'Omnichannel', description: 'Schedule notifications, manage API keys, and monitor status.' },
  omnichannelTest: { title: 'Omnichannel Test', description: 'Run test sends across channels without editing templates.' },
  omnichannelStatus: { title: 'Notification Status', description: 'Runtime delivery status, queue health, and workflow state.' },
  recipients: { title: 'Recipients', description: 'Manage target audiences and recipient groups.' },
  campaigns: { title: 'Campaigns', description: 'Create and manage notification campaigns.' },
  cronJobs: { title: 'Cron Jobs', description: 'Configure recurring notification schedules.' },
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

type ThemeMode = 'system' | 'light' | 'dark';

const THEME_STORAGE_KEY = 'console:theme-mode';
const THEME_OPTIONS: { id: ThemeMode; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: 'light', label: 'Light', icon: Sun },
  { id: 'dark', label: 'Dark', icon: Moon },
  { id: 'system', label: 'System', icon: Monitor },
];

function applyThemeMode(mode: ThemeMode) {
  const root = document.documentElement;
  if (mode === 'system') {
    root.removeAttribute('data-theme');
    return;
  }
  root.setAttribute('data-theme', mode);
}

// ─── AppShell ─────────────────────────────────────────────────────────────────

export function AppShell({ active, children }: { active?: NavSection; children?: React.ReactNode }) {
  const router = useRouter();
  const [projectOpen, setProjectOpen] = useState(false);
  const [projectVisible, setProjectVisible] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const [cmdOpen, setCmdOpen] = useState(false);
  const [cmdVisible, setCmdVisible] = useState(false);
  const [omnichannelOpen, setOmnichannelOpen] = useState(true);
  const [projectSearch, setProjectSearch] = useState('');
  const [cmdSearch, setCmdSearch] = useState('');
  const [selectedProjectId, setSelectedProjectId] = useState(CATALOG_PROJECTS[0]?.id ?? '');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [themeMode, setThemeMode] = useState<ThemeMode>('system');

  const projectBtnRef = useRef<HTMLButtonElement>(null);
  const projectSearchRef = useRef<HTMLInputElement>(null);
  const projectMenuRef = useRef<HTMLDivElement>(null);
  const accountRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const stored = localStorage.getItem('console:selected-project');
    if (stored && CATALOG_PROJECTS.some((p) => p.id === stored)) setSelectedProjectId(stored);

    const collapsedStored = localStorage.getItem('sidebar:collapsed');
    if (collapsedStored !== null) setSidebarCollapsed(collapsedStored === 'true');

    const storedTheme = localStorage.getItem(THEME_STORAGE_KEY);
    if (storedTheme === 'light' || storedTheme === 'dark' || storedTheme === 'system') {
      setThemeMode(storedTheme);
      applyThemeMode(storedTheme);
      return;
    }
    applyThemeMode('system');
  }, []);

  useEffect(() => {
    if (cmdOpen) {
      setCmdVisible(true);
      return;
    }
    const hideTimer = window.setTimeout(() => setCmdVisible(false), 200);
    return () => window.clearTimeout(hideTimer);
  }, [cmdOpen]);

  // Defer cmdOpen by one frame so the palette mounts in closed state and can animate in
  const [cmdOpenDeferred, setCmdOpenDeferred] = useState(false);
  useEffect(() => {
    if (!cmdVisible) {
      setCmdOpenDeferred(false);
      return;
    }
    if (cmdOpen) {
      const raf = requestAnimationFrame(() => setCmdOpenDeferred(true));
      return () => cancelAnimationFrame(raf);
    }
    setCmdOpenDeferred(false);
  }, [cmdVisible, cmdOpen]);

  useEffect(() => {
    if (projectOpen) {
      setProjectVisible(true);
      const focusTimer = window.setTimeout(() => projectSearchRef.current?.focus(), 60);
      return () => window.clearTimeout(focusTimer);
    }
    const hideTimer = window.setTimeout(() => setProjectVisible(false), 220);
    return () => window.clearTimeout(hideTimer);
  }, [projectOpen]);

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
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setCmdOpen((v) => !v);
      }
      if (e.key === 'Escape') {
        setCmdOpen(false);
        setCmdSearch('');
        setProjectOpen(false);
        setAccountOpen(false);
      }
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  function closeCommandPalette() {
    setCmdOpen(false);
    setCmdSearch('');
  }

  function closeProjectMenu() {
    setProjectOpen(false);
    setProjectSearch('');
  }

  function openCommandItem(href: string) {
    closeCommandPalette();
    router.push(href);
  }

  function toggleSidebar() {
    setSidebarCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem('sidebar:collapsed', String(next));
      return next;
    });
  }

  const currentProject = CATALOG_PROJECTS.find((p) => p.id === selectedProjectId);
  const normalizedProjectSearch = projectSearch.trim().toLowerCase();
  const filteredProjects = CATALOG_PROJECTS.filter(
    (p) => p.label.toLowerCase().includes(normalizedProjectSearch) || p.id.toLowerCase().includes(normalizedProjectSearch),
  );
  const filteredShortcuts = PROJECT_SHORTCUTS.filter(
    (item) => !normalizedProjectSearch
      || item.label.toLowerCase().includes(normalizedProjectSearch)
      || item.description.toLowerCase().includes(normalizedProjectSearch),
  );

  function selectProject(id: string) {
    setSelectedProjectId(id);
    localStorage.setItem('console:selected-project', id);
    closeProjectMenu();
  }

  function updateThemeMode(mode: ThemeMode) {
    setThemeMode(mode);
    applyThemeMode(mode);
    if (mode === 'system') {
      localStorage.removeItem(THEME_STORAGE_KEY);
      return;
    }
    localStorage.setItem(THEME_STORAGE_KEY, mode);
  }

  const pageInfo = active ? PAGE_TITLES[active] : undefined;

  function sidebarLink(href: string, label: string, Icon: React.ComponentType<{ className?: string }>, section: NavSection, external = false) {
    const isActive = active === section;
    const baseClass = `flex items-center justify-center rounded-lg py-2 transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`;
    const expandedClass = `flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}`;
    if (external) {
      if (sidebarCollapsed) {
        return <a key={href} href={href} target="_blank" rel="noopener noreferrer" title={label} className={baseClass}><Icon className="h-4 w-4 shrink-0" /></a>;
      }
      return <a key={href} href={href} target="_blank" rel="noopener noreferrer" className={expandedClass}><Icon className="h-4 w-4 shrink-0" />{label}</a>;
    }
    if (sidebarCollapsed) {
      return <Link key={href} href={href} title={label} className={baseClass}><Icon className="h-4 w-4 shrink-0" /></Link>;
    }
    return <Link key={href} href={href} className={expandedClass}><Icon className="h-4 w-4 shrink-0" />{label}</Link>;
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
              className="inline-flex h-8 items-center gap-1.5 rounded-md border border-zinc-700 bg-zinc-900 px-2.5 text-sm text-zinc-100 transition-colors hover:border-zinc-600">
              <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-black ring-1 ring-zinc-700/80">
                <svg viewBox="0 0 24 24" className="h-2.5 w-2.5 fill-white" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                  <path d="M12 4 20 18H4z" />
                </svg>
              </div>
              <span className="max-w-[110px] truncate font-medium">{currentProject?.label ?? 'Select project'}</span>
              <svg className="h-4 w-4 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
              </svg>
            </button>

            {projectVisible && (
              <div className="fixed inset-0 z-[60]">
                <button
                  type="button"
                  aria-label="Close project switcher"
                  onClick={closeProjectMenu}
                  className={`absolute inset-0 transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] ${projectOpen ? 'bg-black/35 opacity-100 backdrop-blur-[1px]' : 'bg-black/0 opacity-0 backdrop-blur-0'}`}
                />
                <div
                  ref={projectMenuRef}
                  className={`absolute left-3 top-[3.35rem] w-[min(430px,calc(100vw-1.5rem))] overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-950/95 shadow-2xl transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] ${projectOpen ? 'translate-y-0 scale-100 opacity-100' : '-translate-y-2 scale-[0.97] opacity-0 pointer-events-none'}`}
                >
                  <div className="flex items-center gap-3 border-b border-zinc-800/90 px-4 py-3">
                    <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-black ring-1 ring-zinc-700/80">
                      <svg viewBox="0 0 24 24" className="h-3 w-3 fill-white" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                        <path d="M12 4 20 18H4z" />
                      </svg>
                    </div>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-zinc-100">Project Switcher</p>
                      <p className="truncate text-xs text-zinc-500">{currentProject?.label ?? 'No project selected'}</p>
                    </div>
                    <span className="ml-auto inline-flex items-center gap-1 rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-400">
                      All Projects
                      <ChevronDown className="h-3 w-3" />
                    </span>
                  </div>

                  <div className="border-b border-zinc-800/90 px-4 py-3">
                    <div className="flex items-center gap-2 rounded-lg border border-zinc-800 bg-zinc-900/80 px-3 py-2">
                      <Search className="h-4 w-4 shrink-0 text-zinc-500" />
                      <input
                        ref={projectSearchRef}
                        type="text"
                        placeholder="Find..."
                        value={projectSearch}
                        onChange={(e) => setProjectSearch(e.target.value)}
                        className="flex-1 bg-transparent text-sm text-zinc-200 placeholder-zinc-500 outline-none"
                      />
                      <kbd className="rounded border border-zinc-700 bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-500">Esc</kbd>
                    </div>
                  </div>

                  <div className="max-h-[min(72vh,560px)] overflow-y-auto p-2">
                    {filteredProjects.length === 0 && filteredShortcuts.length === 0 ? (
                      <p className="px-3 py-8 text-center text-sm text-zinc-500">No results found</p>
                    ) : (
                      <>
                        {filteredProjects.length > 0 && (
                          <div className="mb-2">
                            {filteredProjects.map((project, index) => (
                              <button
                                key={project.id}
                                onClick={() => selectProject(project.id)}
                                className={`w-full rounded-lg px-3 py-2.5 text-left transition-all duration-200 ease-[cubic-bezier(0.16,1,0.3,1)] hover:bg-zinc-900 ${projectOpen ? 'translate-x-0 opacity-100' : 'translate-x-1 opacity-0'}`}
                                style={{ transitionDelay: `${Math.min(index * 12, 120)}ms` }}
                              >
                                <div className="flex items-center gap-3">
                                  <div className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold text-white ${projectColor(project.id)}`}>
                                    {project.label.slice(0, 1).toUpperCase()}
                                  </div>
                                  <div className="min-w-0 flex-1">
                                    <p className="truncate text-[15px] font-medium text-zinc-100">{project.label}</p>
                                    <p className="truncate text-xs text-zinc-500">Project</p>
                                  </div>
                                  {selectedProjectId === project.id ? <Check className="h-4 w-4 shrink-0 text-blue-400" /> : null}
                                </div>
                              </button>
                            ))}
                          </div>
                        )}

                        {filteredShortcuts.length > 0 && (
                          <>
                            <div className="my-1 h-px bg-zinc-800/80" />
                            {filteredShortcuts.map((item, index) => {
                              const Icon = item.icon;
                              return (
                                <Link
                                  key={item.id}
                                  href={item.href}
                                  onClick={closeProjectMenu}
                                  className={`flex items-center gap-3 rounded-lg px-3 py-2.5 transition-all duration-200 ease-[cubic-bezier(0.16,1,0.3,1)] hover:bg-zinc-900 ${projectOpen ? 'translate-x-0 opacity-100' : 'translate-x-1 opacity-0'}`}
                                  style={{ transitionDelay: `${Math.min((filteredProjects.length + index) * 12, 160)}ms` }}
                                >
                                  <Icon className="h-4 w-4 shrink-0 text-zinc-400" />
                                  <div className="min-w-0">
                                    <p className="truncate text-[15px] font-medium text-zinc-100">{item.label}</p>
                                    <p className="truncate text-xs text-zinc-500">{item.description}</p>
                                  </div>
                                </Link>
                              );
                            })}
                          </>
                        )}
                      </>
                    )}
                  </div>

                  <div className="border-t border-zinc-800/90 p-2">
                    <Link href="/projects" onClick={closeProjectMenu} className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-900 hover:text-zinc-100">
                      <Plus className="h-4 w-4" />
                      Create New Project
                    </Link>
                  </div>
                </div>
              </div>
            )}
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
                <div className="px-3 py-2">
                  <div className="flex items-center">
                    <span className="text-sm text-zinc-300">Theme</span>
                    <div className="ml-auto inline-flex items-center rounded-md border border-zinc-700 bg-zinc-900 p-0.5">
                      {THEME_OPTIONS.map((option) => {
                        const Icon = option.icon;
                        const isActive = themeMode === option.id;
                        return (
                          <button
                            key={option.id}
                            type="button"
                            title={option.label}
                            aria-label={`Set ${option.label.toLowerCase()} theme`}
                            onClick={() => updateThemeMode(option.id)}
                            className={`rounded-sm p-1 transition-colors ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-500 hover:text-zinc-300'}`}
                          >
                            <Icon className="h-3.5 w-3.5" />
                          </button>
                        );
                      })}
                    </div>
                  </div>
                </div>
                <div className="my-1 border-t border-zinc-800" />
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

            {MAIN_ITEMS.map((item) => sidebarLink(item.href, item.label, item.icon, item.section, item.external))}

            <div className="my-2 border-t border-zinc-800/70" />

            {sidebarLink('/settings', 'Settings', Settings, 'settings')}
          </nav>

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
      {cmdVisible && (
        <div className="fixed inset-0 z-50">
          <button
            aria-label="Close command palette"
            type="button"
            className={`absolute inset-0 transition-opacity duration-200 ${cmdOpenDeferred ? 'bg-black/60 opacity-100 backdrop-blur-sm' : 'bg-black/40 opacity-0 backdrop-blur-0'}`}
            onClick={closeCommandPalette}
          />
          <div className="relative mx-auto mt-[12vh] max-w-xl px-4">
            <Command
              label="Global command palette"
              loop
              value={cmdSearch}
              onValueChange={setCmdSearch}
              className={`overflow-hidden rounded-xl border border-zinc-700 bg-zinc-950 shadow-2xl transition-all duration-200 ease-out ${cmdOpenDeferred ? 'translate-y-0 scale-100 opacity-100' : '-translate-y-2 scale-[0.98] opacity-0'}`}
            >
              <div className="flex items-center gap-3 border-b border-zinc-800 px-4 py-3">
                <Search className="h-4 w-4 shrink-0 text-zinc-400" />
                <Command.Input
                  autoFocus
                  placeholder="Search pages, projects, actions..."
                  className="flex-1 bg-transparent text-sm text-zinc-100 placeholder-zinc-500 outline-none"
                />
                <button onClick={closeCommandPalette} className="text-zinc-500 transition-colors hover:text-zinc-300">
                  <X className="h-4 w-4" />
                </button>
              </div>
              <Command.List className="max-h-80 overflow-y-auto py-1.5">
                <Command.Empty className="px-3 py-6 text-center text-sm text-zinc-500">No results</Command.Empty>
                {CMD_GROUPS.map((group, groupIndex) => (
                  <div key={group.section}>
                    {groupIndex > 0 ? <Command.Separator className="mx-3 h-px bg-zinc-800/70" /> : null}
                    <Command.Group className="px-1.5 py-1">
                      <p className="px-2 pb-1 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">{group.section}</p>
                      {group.items.map((item) => (
                        <Command.Item
                          key={item.href}
                          value={`${item.label} ${item.section}`}
                          onSelect={() => openCommandItem(item.href)}
                          className={`flex cursor-pointer items-center justify-between rounded-lg px-3 py-2.5 text-sm text-zinc-300 outline-none transition-all duration-200 hover:bg-zinc-900 hover:text-zinc-100 data-[selected=true]:bg-zinc-800 data-[selected=true]:text-zinc-100 ${cmdOpenDeferred ? 'translate-x-0 opacity-100' : 'translate-x-1 opacity-0'}`}
                          style={{ transitionDelay: `${Math.min((CMD_ITEM_ORDER.get(item.href) ?? 0) * 10, 120)}ms` }}
                        >
                          <span>{item.label}</span>
                          <span className="text-xs text-zinc-600">{item.section}</span>
                        </Command.Item>
                      ))}
                    </Command.Group>
                  </div>
                ))}
              </Command.List>
              <div className="flex items-center gap-4 border-t border-zinc-800 px-4 py-2 text-[11px] text-zinc-600">
                <span><kbd className="bg-zinc-800 rounded px-1 text-zinc-500">↵</kbd> Open</span>
                <span><kbd className="bg-zinc-800 rounded px-1 text-zinc-500">Esc</kbd> Close</span>
              </div>
            </Command>
          </div>
        </div>
      )}
    </div>
  );
}

export { AppShell as AppNav };
