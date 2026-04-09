'use client';

import Link from 'next/link';
import { useCallback, useMemo, useState } from 'react';
import { BarChart3, Filter, Plus, CheckCircle2, MessageCircle, ExternalLink, ChevronDown } from 'lucide-react';
import { useAddLeadNote, useCompleteFollowUpTask, useCreateLead, useLeadAnalytics, useLeads, useUpdateLeadStage } from '@/hooks/useLeads';
import { useMyListings } from '@/hooks/useListings';
import { normalizeContactPhone } from '@/lib/listing-contact';
import type { Lead, LeadStage } from '@/types';

const LEAD_SOURCES = [
  'whatsapp',
  'instagram',
  'facebook',
  'tiktok',
  'referral',
  'website',
  'telepon',
  'walk-in',
  'manual',
];

const LEADS_PER_COLUMN = 10;

function waLink(phone: string): string {
  return `https://wa.me/${normalizeContactPhone(phone)}`;
}

const STAGES: Array<{ key: LeadStage; label: string }> = [
  { key: 'new', label: 'Lead baru' },
  { key: 'interested', label: 'Tertarik' },
  { key: 'viewing', label: 'Viewing' },
  { key: 'negotiation', label: 'Negosiasi' },
  { key: 'deal', label: 'Deal' },
  { key: 'lost', label: 'Lost' },
];

export function AgentWorkspaceClient() {
  const [selectedStage, setSelectedStage] = useState<LeadStage | ''>('');
  const [newLeadName, setNewLeadName] = useState('');
  const [newLeadPhone, setNewLeadPhone] = useState('');
  const [newLeadSource, setNewLeadSource] = useState('whatsapp');
  const [newLeadListingId, setNewLeadListingId] = useState('');
  const [noteByLead, setNoteByLead] = useState<Record<string, string>>({});
  const [noteErrorByLead, setNoteErrorByLead] = useState<Record<string, string>>({});
  const [expandedStages, setExpandedStages] = useState<Set<LeadStage>>(new Set());

  const leadQuery = useLeads(selectedStage ? { stage: selectedStage } : undefined);
  const analyticsQuery = useLeadAnalytics();
  const myListingsQuery = useMyListings({ pageSize: 100 });
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

  const toggleExpand = useCallback((stageKey: LeadStage) => {
    setExpandedStages((prev) => {
      const next = new Set(prev);
      if (next.has(stageKey)) next.delete(stageKey);
      else next.add(stageKey);
      return next;
    });
  }, []);

  const submitCreateLead = async () => {
    if (!newLeadName.trim()) return;
    await createLeadMutation.mutateAsync({
      name: newLeadName.trim(),
      phone: newLeadPhone.trim() || undefined,
      source: newLeadSource,
      listingId: newLeadListingId.trim() || undefined,
    });
    setNewLeadName('');
    setNewLeadPhone('');
    setNewLeadListingId('');
  };

  const visibleStages = selectedStage ? STAGES.filter((s) => s.key === selectedStage) : STAGES;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="mb-6 flex items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black text-brand-primary">Agent Operating Tool</h1>
          <p className="mt-1 text-sm text-gray-500">Lead inbox, follow-up, pipeline, dan analytics dalam satu halaman.</p>
        </div>
        <Link href="/listings" className="btn-secondary text-sm">Lihat Iklan Saya</Link>
      </div>

      <div className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard icon={Filter} label="Total Lead" value={analyticsQuery.data?.leadCount ?? 0} />
        <StatCard icon={BarChart3} label="Lead → Viewing" value={`${Math.round((analyticsQuery.data?.leadToViewingRate ?? 0) * 100)}%`} />
        <StatCard icon={CheckCircle2} label="Viewing → Deal" value={`${Math.round((analyticsQuery.data?.viewingToDealRate ?? 0) * 100)}%`} />
      </div>

      <div className="mb-6 rounded-2xl border border-gray-200 bg-white p-4">
        <div className="mb-3 text-sm font-semibold text-gray-800">Tambah lead cepat</div>
        <div className="grid gap-2 md:grid-cols-4">
          <input value={newLeadName} onChange={(e) => setNewLeadName(e.target.value)} placeholder="Nama lead" className="input-field text-sm" />
          <input value={newLeadPhone} onChange={(e) => setNewLeadPhone(e.target.value)} placeholder="Nomor WhatsApp/Telepon" className="input-field text-sm" />
          <select
            value={newLeadSource}
            onChange={(e) => setNewLeadSource(e.target.value)}
            className="input-field text-sm"
          >
            {LEAD_SOURCES.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
          <select
            value={newLeadListingId}
            onChange={(e) => setNewLeadListingId(e.target.value)}
            className="input-field text-sm"
          >
            <option value="">ID Iklan (opsional)</option>
            {(myListingsQuery.data?.items ?? []).map((listing) => (
              <option key={listing.listingId} value={listing.listingId}>
                {listing.title || listing.listingId}
              </option>
            ))}
          </select>
          <button type="button" onClick={submitCreateLead} className="btn-primary flex items-center justify-center gap-2 text-sm md:col-span-4">
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

      <div className={`grid gap-4 ${selectedStage ? 'lg:grid-cols-1' : 'lg:grid-cols-3'}`}>
        {visibleStages.map((stage) => {
          const allLeads = grouped.get(stage.key) ?? [];
          const isExpanded = expandedStages.has(stage.key);
          const visibleLeads = isExpanded ? allLeads : allLeads.slice(0, LEADS_PER_COLUMN);
          const hiddenCount = allLeads.length - visibleLeads.length;

          return (
            <section key={stage.key} className="rounded-2xl border border-gray-200 bg-white p-4">
              <h2 className="mb-3 text-sm font-bold text-gray-800">{stage.label} ({allLeads.length})</h2>
              <div className="space-y-3">
                {visibleLeads.map((lead) => (
                  <LeadCard
                    key={lead.leadId}
                    lead={lead}
                    noteValue={noteByLead[lead.leadId] ?? ''}
                    noteError={noteErrorByLead[lead.leadId] ?? ''}
                    onNoteChange={(val) => {
                      setNoteByLead((prev) => ({ ...prev, [lead.leadId]: val }));
                      if (noteErrorByLead[lead.leadId]) setNoteErrorByLead((prev) => ({ ...prev, [lead.leadId]: '' }));
                    }}
                    onNoteSave={async () => {
                      const note = (noteByLead[lead.leadId] ?? '').trim();
                      if (!note) return;
                      try {
                        await addNoteMutation.mutateAsync({ leadId: lead.leadId, note });
                        setNoteByLead((prev) => ({ ...prev, [lead.leadId]: '' }));
                        setNoteErrorByLead((prev) => ({ ...prev, [lead.leadId]: '' }));
                      } catch {
                        setNoteErrorByLead((prev) => ({ ...prev, [lead.leadId]: 'Gagal menyimpan catatan.' }));
                      }
                    }}
                    onStageChange={(newStage) => updateStageMutation.mutate({ leadId: lead.leadId, stage: newStage })}
                    onCompleteTask={(taskId) => completeFollowUpMutation.mutate({ leadId: lead.leadId, taskId, status: 'completed' })}
                  />
                ))}
                {hiddenCount > 0 && (
                  <button
                    type="button"
                    onClick={() => toggleExpand(stage.key)}
                    className="flex w-full items-center justify-center gap-1 rounded-xl border border-dashed border-gray-300 py-2 text-xs font-semibold text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-700"
                  >
                    <ChevronDown className="h-3.5 w-3.5" />
                    Tampilkan {hiddenCount} lead lainnya
                  </button>
                )}
                {isExpanded && allLeads.length > LEADS_PER_COLUMN && (
                  <button
                    type="button"
                    onClick={() => toggleExpand(stage.key)}
                    className="flex w-full items-center justify-center gap-1 rounded-xl border border-dashed border-gray-300 py-2 text-xs font-semibold text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-700"
                  >
                    Sembunyikan
                  </button>
                )}
                {allLeads.length === 0 && (
                  <p className="rounded-xl border border-dashed border-gray-200 p-3 text-xs text-gray-400">Belum ada lead di tahap ini.</p>
                )}
              </div>
            </section>
          );
        })}
      </div>
    </div>
  );
}

function LeadCard({
  lead,
  noteValue,
  noteError,
  onNoteChange,
  onNoteSave,
  onStageChange,
  onCompleteTask,
}: {
  lead: Lead;
  noteValue: string;
  noteError: string;
  onNoteChange: (val: string) => void;
  onNoteSave: () => void;
  onStageChange: (stage: LeadStage) => void;
  onCompleteTask: (taskId: string) => void;
}) {
  return (
    <article className="rounded-xl border border-gray-100 bg-gray-50 p-3">
      <div className="mb-2 flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-gray-900">{lead.name}</p>
          <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-xs text-gray-500">
            <span className="truncate">{lead.phone || 'Tidak ada nomor'}</span>
            {lead.phone && (
              <a
                href={waLink(lead.phone)}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex flex-shrink-0 items-center gap-0.5 rounded-full bg-[#25D366] px-1.5 py-0.5 text-[10px] font-semibold text-white hover:bg-[#1ebe5d]"
                title="Chat via WhatsApp"
              >
                <MessageCircle className="h-2.5 w-2.5" />
                WA
              </a>
            )}
            <span>•</span>
            <span>{lead.source || 'manual'}</span>
          </div>
          {lead.listingId && (
            <Link
              href={`/listings/${lead.listingId}`}
              className="mt-1 inline-flex items-center gap-1 rounded-full bg-brand-light px-2 py-0.5 text-[10px] font-medium text-brand-primary transition-colors hover:bg-brand-primary hover:text-white"
            >
              <ExternalLink className="h-2.5 w-2.5" />
              Lihat Iklan
            </Link>
          )}
        </div>
        <select
          value={lead.stage}
          onChange={(e) => onStageChange(e.target.value as LeadStage)}
          className="flex-shrink-0 rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs"
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
              onClick={() => onCompleteTask(task.taskId)}
              className="font-semibold text-brand-primary"
            >
              Selesai
            </button>
          </div>
        ))}
      </div>

      {(lead.notes ?? []).length > 0 && (
        <div className="mb-2 max-h-24 space-y-1 overflow-y-auto">
          {(lead.notes ?? []).slice().reverse().map((note, idx) => (
            <p key={idx} className="rounded-lg bg-white px-2 py-1 text-xs text-gray-600">
              {note}
            </p>
          ))}
        </div>
      )}

      <div className="flex gap-2">
        <input
          value={noteValue}
          onChange={(e) => onNoteChange(e.target.value)}
          placeholder="Tambahkan catatan follow-up"
          className="input-field h-9 text-xs"
        />
        <button
          type="button"
          className="btn-secondary h-9 px-3 text-xs"
          onClick={onNoteSave}
        >
          Simpan
        </button>
      </div>
      {noteError && (
        <p className="mt-1 text-[10px] text-red-500">{noteError}</p>
      )}
    </article>
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
