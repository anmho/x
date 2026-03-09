'use client';

import { useEffect, useState } from 'react';
import { CreateTemplateRequest, templateApi, Template } from '@/lib/api';
import { Eye, FileText, Pencil, Plus, Send, Trash2, X } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';

type TemplateFormData = {
  name: string;
  description: string;
  subject: string;
  body: string;
  variables: string;
};

type TemplateEditorTab = 'code' | 'preview';

function toErrorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

const EMPTY_FORM: TemplateFormData = {
  name: '',
  description: '',
  subject: '',
  body: '',
  variables: '',
};

const INITIAL_FAKE_TEMPLATES: Template[] = [
  {
    id: 'fake-template-1',
    name: 'Welcome Email',
    description: 'Welcome email for new users',
    subject: 'Welcome to {{app_name}}, {{user_name}}!',
    body: '<html><body><h1>Welcome {{user_name}}!</h1><p>Thank you for joining {{app_name}}.</p></body></html>',
    variables: ['user_name', 'app_name'],
    is_active: true,
    created_at: '2026-03-01T12:00:00.000Z',
    updated_at: '2026-03-01T12:00:00.000Z',
  },
  {
    id: 'fake-template-2',
    name: 'Password Reset',
    description: 'Password reset request email',
    subject: 'Reset your password',
    body: '<html><body><h1>Password Reset Request</h1><p>Hi {{user_name}}, click <a href="{{reset_url}}">here</a> to reset.</p></body></html>',
    variables: ['user_name', 'reset_url'],
    is_active: true,
    created_at: '2026-03-02T12:00:00.000Z',
    updated_at: '2026-03-02T12:00:00.000Z',
  },
];

export default function Templates() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [fakeTemplates, setFakeTemplates] = useState<Template[]>(INITIAL_FAKE_TEMPLATES);
  const [useFakeData, setUseFakeData] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [createFormData, setCreateFormData] = useState<TemplateFormData>(EMPTY_FORM);
  const [editFormData, setEditFormData] = useState<TemplateFormData>(EMPTY_FORM);
  const [autoFakeReason, setAutoFakeReason] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
  const isAnyModalOpen = showCreateModal || showEditModal || showViewModal || deleteTargetId !== null;

  useEffect(() => {
    if (process.env.NEXT_PUBLIC_API_URL) return;
    if (typeof window === 'undefined') return;
    const hostname = window.location.hostname;
    const isLocal = hostname === 'localhost' || hostname === '127.0.0.1';
    if (!isLocal) {
      setUseFakeData(true);
      setAutoFakeReason('Using fake data by default because NEXT_PUBLIC_API_URL is not set for this environment.');
    }
  }, []);

  useEffect(() => {
    if (useFakeData) {
      setTemplates(fakeTemplates);
      setLoading(false);
      setError(null);
      return;
    }
    void loadTemplates();
  }, [useFakeData, fakeTemplates]);

  useEffect(() => {
    if (!isAnyModalOpen) return;

    function onKeyDown(event: KeyboardEvent) {
      if (event.key !== 'Escape') return;
      event.preventDefault();

      if (deleteTargetId !== null) {
        setDeleteTargetId(null);
        return;
      }
      if (showEditModal) {
        setShowEditModal(false);
        return;
      }
      if (showViewModal) {
        setShowViewModal(false);
        return;
      }
      if (showCreateModal) {
        setShowCreateModal(false);
      }
    }

    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [isAnyModalOpen, showCreateModal, showEditModal, showViewModal, deleteTargetId]);

  async function loadTemplates() {
    try {
      setLoading(true);
      const response = await templateApi.list();
      const normalized = (response.data || []).map((template) => normalizeTemplate(template));
      setTemplates(normalized);
      setError(null);
    } catch (err: unknown) {
      setError(toErrorMessage(err) || 'Failed to load templates');
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      const payload: CreateTemplateRequest = {
        name: createFormData.name,
        description: createFormData.description,
        subject: createFormData.subject,
        body: createFormData.body,
        variables: parseVariables(createFormData.variables),
      };

      if (useFakeData) {
        const now = new Date().toISOString();
        const newTemplate: Template = {
          id: `fake-template-${Date.now()}`,
          ...payload,
          is_active: true,
          created_at: now,
          updated_at: now,
        };
        setFakeTemplates((prev) => [newTemplate, ...prev]);
      } else {
        await templateApi.create(payload);
        await loadTemplates();
      }

      setShowCreateModal(false);
      setCreateFormData(EMPTY_FORM);
    } catch (err: unknown) {
      setActionError(`Failed to create template: ${toErrorMessage(err)}`);
    }
  }

  function handleDelete(id: string) {
    setDeleteTargetId(id);
  }

  async function confirmDelete() {
    if (deleteTargetId === null) return;
    const id = deleteTargetId;
    setDeleteTargetId(null);
    try {
      if (useFakeData) {
        setFakeTemplates((prev) => prev.filter((template) => template.id !== id));
      } else {
        await templateApi.delete(id);
        await loadTemplates();
      }
    } catch (err: unknown) {
      setActionError(`Failed to delete template: ${toErrorMessage(err)}`);
    }
  }

  async function handleView(template: Template) {
    try {
      if (useFakeData) {
        setSelectedTemplate(template);
      } else {
        const loaded = await templateApi.get(template.id);
        setSelectedTemplate(normalizeTemplate(loaded));
      }
      setShowViewModal(true);
    } catch (err: unknown) {
      setActionError(`Failed to load template details: ${toErrorMessage(err)}`);
    }
  }

  function handleStartEdit(template: Template) {
    setSelectedTemplate(template);
    setEditFormData({
      name: template.name,
      description: template.description || '',
      subject: template.subject,
      body: template.body,
      variables: (template.variables || []).join(', '),
    });
    setShowEditModal(true);
  }

  async function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedTemplate) return;

    const payload: Partial<CreateTemplateRequest> = {
      name: editFormData.name,
      description: editFormData.description,
      subject: editFormData.subject,
      body: editFormData.body,
      variables: parseVariables(editFormData.variables),
    };

    try {
      if (useFakeData) {
        const now = new Date().toISOString();
        const updated: Template = normalizeTemplate({
          ...selectedTemplate,
          ...payload,
          updated_at: now,
        } as Template);
        setFakeTemplates((prev) => prev.map((template) => (template.id === selectedTemplate.id ? updated : template)));
        setSelectedTemplate(updated);
      } else {
        const updated = await templateApi.update(selectedTemplate.id, payload);
        setSelectedTemplate(normalizeTemplate(updated));
        await loadTemplates();
      }
      setShowEditModal(false);
    } catch (err: unknown) {
      setActionError(`Failed to update template: ${toErrorMessage(err)}`);
    }
  }

  function handleToggleFakeData() {
    setUseFakeData((prev) => !prev);
    setShowCreateModal(false);
    setShowEditModal(false);
    setShowViewModal(false);
    setSelectedTemplate(null);
    setError(null);
  }

  return (
    <AppNav active="templates">
      <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <div className="mb-6 flex items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Templates</h1>
            <p className="mt-1 text-sm text-zinc-400">Reusable notification content blocks.</p>
          </div>
          <div className="flex flex-wrap items-center justify-end gap-3">
            <label className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300">
              <input type="checkbox" checked={useFakeData} onChange={handleToggleFakeData} className="h-4 w-4 rounded border-zinc-600 bg-zinc-800 text-zinc-200" />
              Use Fake Data
            </label>
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-medium hover:bg-zinc-800"
            >
              <Plus className="h-4 w-4" />
              Create Template
            </button>
          </div>
        </div>

        {useFakeData && (
          <p className="mb-4 rounded-md border border-amber-400/30 bg-amber-400/10 px-3 py-2 text-sm text-amber-200">
            Fake data mode is enabled. Create, edit, and delete changes are local to this page session.
            {autoFakeReason ? ` ${autoFakeReason}` : ''}
          </p>
        )}

        {error && (
          <div className="mb-4 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">
            <p>{error}</p>
            {!useFakeData && (
              <button
                type="button"
                onClick={() => setUseFakeData(true)}
                className="mt-2 rounded-md border border-rose-400/40 bg-rose-500/10 px-2.5 py-1 text-xs text-rose-200 hover:bg-rose-500/20"
              >
                Switch to Fake Data
              </button>
            )}
          </div>
        )}

        {actionError && (
          <div className="mb-4 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300 flex items-start justify-between gap-2">
            <p>{actionError}</p>
            <button type="button" onClick={() => setActionError(null)} className="shrink-0 text-rose-400 hover:text-rose-200">
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {loading ? (
          <div className="flex h-64 items-center justify-center">
            <div className="h-10 w-10 animate-spin rounded-full border-b-2 border-zinc-300" />
          </div>
        ) : templates.length === 0 ? (
          <div className="rounded-xl border border-zinc-800 bg-zinc-950 py-14 text-center">
            <FileText className="mx-auto h-10 w-10 text-zinc-600" />
            <h3 className="mt-3 text-lg font-medium">No templates yet</h3>
            <p className="mt-1 text-sm text-zinc-400">Create one to speed up sending.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {templates.map((template) => (
              <article key={template.id} className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
                <div className="mb-3 flex items-start justify-between gap-2">
                  <div>
                    <h3 className="font-medium">{template.name}</h3>
                    <p className="text-sm text-zinc-400">{template.description}</p>
                  </div>
                  <span className={`rounded-full px-2 py-0.5 text-xs ${template.is_active ? 'bg-emerald-500/10 text-emerald-300' : 'bg-zinc-800 text-zinc-400'}`}>
                    {template.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>

                <p className="text-xs uppercase tracking-wide text-zinc-500">Subject</p>
                <p className="mb-3 mt-1 truncate text-sm text-zinc-300">{template.subject}</p>

                {template.variables.length > 0 && (
                  <div className="mb-3 flex flex-wrap gap-1.5">
                    {template.variables.map((variable) => (
                      <span key={variable} className="rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-300">
                        {`{{${variable}}}`}
                      </span>
                    ))}
                  </div>
                )}

                <div className="flex justify-end gap-1 border-t border-zinc-800 pt-3">
                  <button
                    onClick={() => void handleView(template)}
                    className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
                    title="View template"
                  >
                    <Eye className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleStartEdit(template)}
                    className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
                    title="Edit template"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button onClick={() => handleDelete(template.id)} className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-rose-300">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </main>

      {showCreateModal && (
        <ModalShell title="Create Template" onClose={() => setShowCreateModal(false)}>
          <form onSubmit={handleSubmit} className="mt-4 space-y-3 text-sm">
            <TemplateFormFields formData={createFormData} onChange={setCreateFormData} />
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setShowCreateModal(false)} className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2">
                Cancel
              </button>
              <button type="submit" className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-medium hover:bg-zinc-800">
                Create
              </button>
            </div>
          </form>
        </ModalShell>
      )}

      {showEditModal && selectedTemplate && (
        <TemplateEditor
          template={selectedTemplate}
          formData={editFormData}
          onChange={setEditFormData}
          onSubmit={handleEditSubmit}
          onClose={() => setShowEditModal(false)}
        />
      )}

      {deleteTargetId !== null && (
        <ModalShell title="Delete Template" onClose={() => setDeleteTargetId(null)}>
          <p className="mt-3 text-sm text-zinc-300">Are you sure you want to delete this template? This cannot be undone.</p>
          <div className="flex justify-end gap-2 pt-4">
            <button type="button" onClick={() => setDeleteTargetId(null)} className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm">
              Cancel
            </button>
            <button type="button" onClick={() => void confirmDelete()} className="rounded-md border border-rose-700 bg-rose-900/40 px-3 py-2 text-sm font-medium text-rose-200 hover:bg-rose-800/40">
              Delete
            </button>
          </div>
        </ModalShell>
      )}

      {showViewModal && selectedTemplate && (
        <ModalShell title={selectedTemplate.name} maxWidthClassName="max-w-3xl" onClose={() => setShowViewModal(false)}>
          <div className="space-y-4 text-sm">
            <div>
              <p className="text-xs uppercase tracking-wide text-zinc-500">Description</p>
              <p className="mt-1 text-zinc-300">{selectedTemplate.description || 'No description'}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-zinc-500">Subject</p>
              <p className="mt-1 text-zinc-300">{selectedTemplate.subject}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-zinc-500">Variables</p>
              <div className="mt-1 flex flex-wrap gap-1.5">
                {selectedTemplate.variables.length > 0 ? (
                  selectedTemplate.variables.map((variable) => (
                    <span key={variable} className="rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-300">
                      {`{{${variable}}}`}
                    </span>
                  ))
                ) : (
                  <span className="text-zinc-500">No variables</span>
                )}
              </div>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-zinc-500">Body (HTML)</p>
              <pre className="mt-1 max-h-64 overflow-auto whitespace-pre-wrap rounded-md border border-zinc-800 bg-zinc-900 p-3 font-mono text-xs text-zinc-200">
                {selectedTemplate.body}
              </pre>
            </div>
            <div className="flex justify-end gap-2 border-t border-zinc-800 pt-3">
              <button
                type="button"
                onClick={() => {
                  setShowViewModal(false);
                  handleStartEdit(selectedTemplate);
                }}
                className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2"
              >
                Edit
              </button>
              <button type="button" onClick={() => setShowViewModal(false)} className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2">
                Close
              </button>
            </div>
          </div>
        </ModalShell>
      )}
    </AppNav>
  );
}

function parseVariables(input: string): string[] {
  return Array.from(
    new Set(
      input
        .split(',')
        .map((value) => value.trim())
        .filter((value) => value.length > 0)
    )
  );
}

function normalizeTemplate(template: Template): Template {
  return {
    ...template,
    description: template.description || '',
    body: template.body || '',
    variables: Array.isArray(template.variables) ? template.variables : [],
  };
}

function TemplateFormFields({
  formData,
  onChange,
}: {
  formData: TemplateFormData;
  onChange: React.Dispatch<React.SetStateAction<TemplateFormData>>;
}) {
  return (
    <>
      <Field label="Template Name">
        <input required value={formData.name} onChange={(e) => onChange((prev) => ({ ...prev, name: e.target.value }))} className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2" />
      </Field>
      <Field label="Description">
        <input value={formData.description} onChange={(e) => onChange((prev) => ({ ...prev, description: e.target.value }))} className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2" />
      </Field>
      <Field label="Subject">
        <input required value={formData.subject} onChange={(e) => onChange((prev) => ({ ...prev, subject: e.target.value }))} className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2" />
      </Field>
      <Field label="Message Body (HTML)">
        <textarea required rows={8} value={formData.body} onChange={(e) => onChange((prev) => ({ ...prev, body: e.target.value }))} className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-mono" />
      </Field>
      <Field label="Variables (comma-separated)">
        <input value={formData.variables} onChange={(e) => onChange((prev) => ({ ...prev, variables: e.target.value }))} className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2" />
      </Field>
    </>
  );
}

function ModalShell({
  title,
  children,
  maxWidthClassName = 'max-w-2xl',
  onClose,
}: {
  title: string;
  children: React.ReactNode;
  maxWidthClassName?: string;
  onClose?: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 animate-[modal-overlay-in_140ms_ease-out]"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose?.();
      }}
    >
      <div
        className={`max-h-[90vh] w-full ${maxWidthClassName} overflow-y-auto rounded-xl border border-zinc-800 bg-zinc-950 p-5 shadow-2xl animate-[modal-content-in_180ms_ease-out]`}
        onMouseDown={(event) => event.stopPropagation()}
      >
        <h3 className="text-xl font-semibold">{title}</h3>
        {children}
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-zinc-400">{label}</span>
      {children}
    </label>
  );
}

function TemplateEditor({
  template,
  formData,
  onChange,
  onSubmit,
  onClose,
}: {
  template: Template;
  formData: TemplateFormData;
  onChange: React.Dispatch<React.SetStateAction<TemplateFormData>>;
  onSubmit: (e: React.FormEvent) => void;
  onClose: () => void;
}) {
  const [tab, setTab] = useState<TemplateEditorTab>('code');
  const [testEmail, setTestEmail] = useState('');
  const [testSending, setTestSending] = useState(false);
  const [testResult, setTestResult] = useState<{ ok: boolean; msg: string } | null>(null);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key !== 'Escape') return;
      event.preventDefault();
      onClose();
    }

    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [onClose]);

  async function handleTestSend(e: React.FormEvent) {
    e.preventDefault();
    if (!testEmail.trim()) return;
    setTestSending(true);
    setTestResult(null);
    try {
      const { notificationApi } = await import('@/lib/api');
      await notificationApi.create({
        channel: 'email',
        recipient: testEmail.trim(),
        recipient_email: testEmail.trim(),
        subject: formData.subject || template.subject,
        body: formData.body || template.body,
        metadata: { source: 'template-preview', template_id: template.id },
      });
      setTestResult({ ok: true, msg: `Test sent to ${testEmail}` });
    } catch (err: unknown) {
      setTestResult({ ok: false, msg: toErrorMessage(err) || 'Send failed' });
    } finally {
      setTestSending(false);
    }
  }

  const previewHtml = formData.body || '<p class="text-zinc-500 text-sm">Start typing HTML in the code editor…</p>';

  return (
    <div className="fixed inset-0 z-50 flex flex-col app-shell-bg animate-[modal-overlay-in_120ms_ease-out]">
      {/* Header */}
      <div className="flex h-14 items-center justify-between border-b border-zinc-800 px-5">
        <div className="flex items-center gap-3">
          <button
            onClick={onClose}
            className="flex items-center justify-center rounded-md p-1.5 text-zinc-500 hover:bg-zinc-900 hover:text-zinc-300 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
          <div>
            <p className="text-sm font-semibold text-zinc-100">{formData.name || template.name}</p>
            <p className="text-xs text-zinc-500">{formData.subject || template.subject}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Tab switcher */}
          <div className="flex items-center gap-0.5 rounded-lg border border-zinc-800 bg-zinc-900 p-0.5">
            {(['code', 'preview'] as TemplateEditorTab[]).map((t) => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`rounded-md px-3 py-1 text-xs font-medium capitalize transition-colors ${
                  tab === t ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-500 hover:text-zinc-300'
                }`}
              >
                {t === 'code' ? 'Code' : 'Preview'}
              </button>
            ))}
          </div>

          <form onSubmit={onSubmit} className="flex items-center gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:bg-zinc-800 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-500 transition-colors"
            >
              Save changes
            </button>
          </form>
        </div>
      </div>

      {/* Body: split-pane */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: form fields */}
        <div className="flex w-[380px] shrink-0 flex-col overflow-y-auto border-r border-zinc-800 bg-zinc-950">
          <form onSubmit={onSubmit} className="space-y-4 p-4 text-sm">
            <Field label="Name">
              <input
                required
                value={formData.name}
                onChange={(e) => onChange((p) => ({ ...p, name: e.target.value }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
              />
            </Field>
            <Field label="Description">
              <input
                value={formData.description}
                onChange={(e) => onChange((p) => ({ ...p, description: e.target.value }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
              />
            </Field>
            <Field label="Subject">
              <input
                required
                value={formData.subject}
                onChange={(e) => onChange((p) => ({ ...p, subject: e.target.value }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
              />
            </Field>
            <Field label="Variables (comma-separated)">
              <input
                value={formData.variables}
                onChange={(e) => onChange((p) => ({ ...p, variables: e.target.value }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
                placeholder="user_name, app_name"
              />
            </Field>
          </form>

          {/* Test send */}
          <div className="border-t border-zinc-800 p-4">
            <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-zinc-500">Test Send</p>
            <form onSubmit={handleTestSend} className="space-y-2">
              <input
                type="email"
                required
                value={testEmail}
                onChange={(e) => setTestEmail(e.target.value)}
                placeholder="recipient@example.com"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
              />
              <button
                type="submit"
                disabled={testSending || !testEmail.trim()}
                className="inline-flex w-full items-center justify-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-medium text-zinc-200 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 transition-colors"
              >
                <Send className="h-3.5 w-3.5" />
                {testSending ? 'Sending…' : 'Send test email'}
              </button>
            </form>
            {testResult && (
              <p className={`mt-2 rounded-md border px-3 py-2 text-xs ${testResult.ok ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300' : 'border-rose-500/30 bg-rose-500/10 text-rose-300'}`}>
                {testResult.msg}
              </p>
            )}
          </div>
        </div>

        {/* Right: code editor or preview */}
        <div className="flex flex-1 flex-col overflow-hidden">
          {tab === 'code' ? (
            <textarea
              value={formData.body}
              onChange={(e) => onChange((p) => ({ ...p, body: e.target.value }))}
              spellCheck={false}
              className="flex-1 resize-none bg-[#111] p-5 font-mono text-sm text-zinc-200 focus:outline-none"
              placeholder="<html><body><!-- your email HTML --></body></html>"
            />
          ) : (
            <div className="flex flex-1 flex-col overflow-hidden">
              {/* Preview toolbar */}
              <div className="flex items-center gap-2 border-b border-zinc-800 bg-zinc-950 px-4 py-2">
                <div className="flex gap-1.5">
                  <div className="h-3 w-3 rounded-full bg-zinc-700" />
                  <div className="h-3 w-3 rounded-full bg-zinc-700" />
                  <div className="h-3 w-3 rounded-full bg-zinc-700" />
                </div>
                <p className="flex-1 text-center text-xs text-zinc-600">Email Preview</p>
              </div>
              <iframe
                srcDoc={previewHtml}
                sandbox="allow-same-origin"
                className="flex-1 bg-white"
                title="Template preview"
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
