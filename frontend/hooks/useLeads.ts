'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addLeadNote,
  completeFollowUpTask,
  createLead,
  getLead,
  getLeadAnalytics,
  getLeads,
  updateLeadStage,
} from '@/lib/api';
import type {
  AddLeadNoteRequest,
  CompleteFollowUpTaskRequest,
  CreateLeadRequest,
  FollowUpTaskStatus,
  LeadStage,
} from '@/types';

export function useLeads(params?: { stage?: LeadStage; dueOnly?: boolean; limit?: number; cursor?: string }) {
  return useQuery({
    queryKey: ['leads', params ?? {}],
    queryFn: () => getLeads(params),
  });
}

export function useLead(leadId: string) {
  return useQuery({
    queryKey: ['lead', leadId],
    queryFn: () => getLead(leadId),
    enabled: Boolean(leadId),
  });
}

export function useLeadAnalytics() {
  return useQuery({
    queryKey: ['lead-analytics'],
    queryFn: () => getLeadAnalytics(),
  });
}

export function useCreateLead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateLeadRequest) => createLead(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leads'] });
      queryClient.invalidateQueries({ queryKey: ['lead-analytics'] });
    },
  });
}

export function useUpdateLeadStage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ leadId, stage, reason }: { leadId: string; stage: LeadStage; reason?: string }) =>
      updateLeadStage(leadId, { stage, reason }),
    onSuccess: (lead) => {
      queryClient.setQueryData(['lead', lead.leadId], lead);
      queryClient.invalidateQueries({ queryKey: ['leads'] });
      queryClient.invalidateQueries({ queryKey: ['lead-analytics'] });
    },
  });
}

export function useAddLeadNote() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ leadId, note }: { leadId: string; note: string }) =>
      addLeadNote(leadId, { note } as AddLeadNoteRequest),
    onSuccess: (lead) => {
      queryClient.setQueryData(['lead', lead.leadId], lead);
      queryClient.invalidateQueries({ queryKey: ['leads'] });
    },
  });
}

export function useCompleteFollowUpTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      leadId,
      taskId,
      status,
      note,
    }: {
      leadId: string;
      taskId: string;
      status: FollowUpTaskStatus;
      note?: string;
    }) =>
      completeFollowUpTask(leadId, taskId, {
        status,
        note,
      } as CompleteFollowUpTaskRequest),
    onSuccess: (lead) => {
      queryClient.setQueryData(['lead', lead.leadId], lead);
      queryClient.invalidateQueries({ queryKey: ['leads'] });
      queryClient.invalidateQueries({ queryKey: ['lead-analytics'] });
    },
  });
}
