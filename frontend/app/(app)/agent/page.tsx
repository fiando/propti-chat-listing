'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { BarChart3, Clock3, Filter, Plus, CheckCircle2 } from 'lucide-react';
import { useAddLeadNote, useCompleteFollowUpTask, useCreateLead, useLeadAnalytics, useLeads, useUpdateLeadStage } from '@/hooks/useLeads';
import type { Lead, LeadStage } from '@/types';

const STAGES: Array<{ key: LeadStage; label: string }> = [
  { key: 'new', label: 'Lead baru' },
  { key: 'interested', label: 'Tertarik' },
  { key: 'viewing', label: 'Viewing' },
  { key: 'negotiation', label: 'Negosiasi' },
  { key: 'deal', label: 'Deal' },
  { key: 'lost', label: 'Lost' },
];

export default function AgentWorkspacePage() {
  const [selectedStage, setSelectedStage] = useState<LeadStage | ''>('');
  const [newLeadName, setNewLeadName] = useState('');
  const [newLeadPhone, setNewLeadPhone] = useState('');
  const [newLeadSource, setNewLeadSource] = useState('whatsapp');
  const [noteByLead, setNoteByLead] = useState<Record<string, string>>({});

  const leadQuery = useLeads(selectedStage ? { stage: selectedStage } : undefined);
  const analyticsQuery = useLeadAnalytics();
  const createLeadMutation = useCreateLead();
  const updateStageMutation = useUpdateLeadStage();
  const addNoteMutation = useAddLeadNote();
  const completeFollowUpMutation = useCompleteFollowUpTask();

  const grouped = useMemo(() => {
    const leads = leadQuery.data?.leads ?? [];
    const map = new Map<LeadStage, Lead[]>();
    for (const stage of STAGES) map.set(stage.key, []);
    for (const lead of leads) {
      const bucket = map.get(lead.stage);
      if (bucket) bucket.push(lead);
    }
    return map;
  }, [leadQuery.data?.leads]);

  const submitCreateLead = async () => {
    if (!newLeadName.trim()) return;
    await createLeadMutation.mutateAsync({
      name: newLeadName.trim(),
      phone: newLeadPhone.trim() || undefined,
      source: newLeadSource.trim() || 'manual',
    });
    setNewLeadName('');
    setNewLeadPhone('');
  };

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="mb-6 flex items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black text-brand-primary">Agent Operating Tool</h1>
          <p className="mt-1 text-sm text-gray-500">Lead inbox, follow-up, pipeline, dan analytics dalam satu halaman.</p>
        </div>
        <Link href="/listings" className="btn-secondary text-sm">Lihat Iklan Saya</Link>
      </div>

      <div className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={Filter} label="Total Lead" value={analyticsQuery.data?.leadCount ?? 0} />
        <StatCard icon={Clock3} label="Median Response (menit)" value={analyticsQuery.data?.medianResponseMinutes ?? 0} />
        <StatCard icon={BarChart3} label="Lead → Viewing" value={`${Math.round((analyticsQuery.data?.leadToViewingRate ?? 0) * 100)}%`} />
        <StatCard icon={CheckCircle2} label="Viewing → Deal" value={`${Math.round((analyticsQuery.data?.viewingToDealRate ?? 0) * 100)}%`} />
      </div>

      <div className="mb-6 rounded-2xl border border-gray-200 bg-white p-4">
        <div className="mb-3 text-sm font-semibold text-gray-800">Tambah lead cepat</div>
        <div className="grid gap-2 md:grid-cols-4">
          <input value={newLeadName} onChange={(e) => setNewLeadName(e.target.value)} placeholder="Nama lead" className="input-field text-sm" />
          <input value={newLeadPhone} onChange={(e) => setNewLeadPhone(e.target.value)} placeholder="Nomor WhatsApp/Telepon" className="input-field text-sm" />
          <input value={newLeadSource} onChange={(e) => setNewLeadSource(e.target.value)} placeholder="Sumber (mis. whatsapp)" className="input-field text-sm" />
          <button type="button" onClick={submitCreateLead} className="btn-primary flex items-center justify-center gap-2 text-sm">
            <Plus className="h-4 w-4" /> Tambah Lead
          </button>
        </div>
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <button type="button" onClick={() => setSelectedStage('')} className={`rounded-full px-3 py-1 text-xs font-semibold ${selectedStage === '' ? 'bg-brand-primary text-white' : 'bg-gray-100 text-gray-600'}`}>Semua</button>
        {STAGES.map((stage) => (
          <button key={stage.key} type="button" onClick={() => setSelectedStage(stage.key)} className={`rounded-full px-3 py-1 text-xs font-semibold ${selectedStage === stage.key ? 'bg-brand-primary text-white' : 'bg-gray-100 text-gray-600'}`}>
            {stage.label}
          </button>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {STAGES.map((stage) => (
          <section key={stage.key} className="rounded-2xl border border-gray-200 bg-white p-4">
            <h2 className="mb-3 text-sm font-bold text-gray-800">{stage.label} ({grouped.get(stage.key)?.length ?? 0})</h2>
            <div className="space-y-3">
              {(grouped.get(stage.key) ?? []).map((lead) => (
                <article key={lead.leadId} className="rounded-xl border border-gray-100 bg-gray-50 p-3">
                  <div className="mb-2 flex items-start justify-between gap-2">
                    <div>
                      <p className="text-sm font-semibold text-gray-900">{lead.name}</p>
                      <p className="text-xs text-gray-500">{lead.phone || '-'} • {lead.source || 'manual'}</p>
                    </div>
                    <select
                      value={lead.stage}
                      onChange={(e) => updateStageMutation.mutate({ leadId: lead.leadId, stage: e.target.value as LeadStage })}
                      className="rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs"
                    >
                      {STAGES.map((s) => <option key={s.key} value={s.key}>{s.label}</option>)}
                    </select>
                  </div>

                  <div className="mb-2 space-y-1">
                    {(lead.followUpTasks ?? []).filter((t) => t.status === 'pending').slice(0, 2).map((task) => (
                      <div key={task.taskId} className="flex items-center justify-between rounded-lg bg-white px-2 py-1 text-xs text-gray-600">
                        <span>Follow-up H+{task.offsetDays} • {new Date(task.dueAt).toLocaleDateString('id-ID')}</span>
                        <button
                          type="button"
                          onClick={() => completeFollowUpMutation.mutate({ leadId: lead.leadId, taskId: task.taskId, status: 'completed' })}
                          className="font-semibold text-brand-primary"
                        >
                          Selesai
                        </button>
                      </div>
                    ))}
                  </div>

                  <div className="flex gap-2">
                    <input
                      value={noteByLead[lead.leadId] ?? ''}
                      onChange={(e) => setNoteByLead((prev) => ({ ...prev, [lead.leadId]: e.target.value }))}
                      placeholder="Tambahkan catatan follow-up"
                      className="input-field h-9 text-xs"
                    />
                    <button
                      type="button"
                      className="btn-secondary h-9 px-3 text-xs"
                      onClick={async () => {
                        const note = (noteByLead[lead.leadId] ?? '').trim();
                        if (!note) return;
                        await addNoteMutation.mutateAsync({ leadId: lead.leadId, note });
                        setNoteByLead((prev) => ({ ...prev, [lead.leadId]: '' }));
                      }}
                    >
                      Simpan
                    </button>
                  </div>
                </article>
              ))}
              {(grouped.get(stage.key) ?? []).length === 0 && (
                <p className="rounded-xl border border-dashed border-gray-200 p-3 text-xs text-gray-400">Belum ada lead di tahap ini.</p>
              )}
            </div>
          </section>
        ))}
      </div>
    </div>
  );
}

function StatCard({ icon: Icon, label, value }: { icon: React.ComponentType<{ className?: string }>; label: string; value: string | number }) {
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-4">
      <div className="mb-2 flex h-10 w-10 items-center justify-center rounded-xl bg-brand-light">
        <Icon className="h-5 w-5 text-brand-primary" />
      </div>
      <div className="text-lg font-black text-gray-900">{value}</div>
      <div className="text-xs text-gray-500">{label}</div>
    </div>
  );
}
