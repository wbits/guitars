import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from '@tanstack/react-query';
import {
  guitarListSchema,
  guitarSchema,
  type Guitar,
  type GuitarInput,
} from '@/domain/guitar';
import { apiFetch } from './client';

const QUERY_KEYS = {
  all: ['guitars'] as const,
  list: () => [...QUERY_KEYS.all, 'list'] as const,
  detail: (id: string) => [...QUERY_KEYS.all, 'detail', id] as const,
};

export const listGuitars = async (signal?: AbortSignal): Promise<Guitar[]> => {
  const raw = await apiFetch<unknown>({ path: '/guitar', signal });
  return guitarListSchema.parse(raw);
};

export const getGuitar = async (id: string, signal?: AbortSignal): Promise<Guitar> => {
  const raw = await apiFetch<unknown>({ path: `/guitar/${encodeURIComponent(id)}`, signal });
  return guitarSchema.parse(raw);
};

export const createGuitar = async (input: GuitarInput): Promise<Guitar> => {
  const raw = await apiFetch<unknown>({ method: 'POST', path: '/guitar', body: input });
  return guitarSchema.parse(raw);
};

export const updateGuitar = async (id: string, input: GuitarInput): Promise<Guitar> => {
  const raw = await apiFetch<unknown>({
    method: 'PUT',
    path: `/guitar/${encodeURIComponent(id)}`,
    body: input,
  });
  return guitarSchema.parse(raw);
};

export const deleteGuitar = async (id: string): Promise<void> => {
  await apiFetch<void>({ method: 'DELETE', path: `/guitar/${encodeURIComponent(id)}` });
};

export const useGuitars = (): UseQueryResult<Guitar[]> =>
  useQuery({
    queryKey: QUERY_KEYS.list(),
    queryFn: ({ signal }) => listGuitars(signal),
  });

export const useGuitar = (id: string | undefined): UseQueryResult<Guitar> =>
  useQuery({
    queryKey: QUERY_KEYS.detail(id ?? ''),
    queryFn: ({ signal }) => getGuitar(id as string, signal),
    enabled: Boolean(id),
  });

export const useCreateGuitar = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: GuitarInput) => createGuitar(input),
    onSuccess: (created) => {
      qc.invalidateQueries({ queryKey: QUERY_KEYS.list() });
      qc.setQueryData(QUERY_KEYS.detail(created.id), created);
    },
  });
};

export const useUpdateGuitar = (id: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: GuitarInput) => updateGuitar(id, input),
    onSuccess: (updated) => {
      qc.invalidateQueries({ queryKey: QUERY_KEYS.list() });
      qc.setQueryData(QUERY_KEYS.detail(updated.id), updated);
    },
  });
};

export const useDeleteGuitar = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteGuitar(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: QUERY_KEYS.list() });
      const previous = qc.getQueryData<Guitar[]>(QUERY_KEYS.list());
      if (previous) {
        qc.setQueryData<Guitar[]>(
          QUERY_KEYS.list(),
          previous.filter((g) => g.id !== id),
        );
      }
      return { previous };
    },
    onError: (_err, _id, context) => {
      if (context?.previous) {
        qc.setQueryData(QUERY_KEYS.list(), context.previous);
      }
    },
    onSettled: (_data, _err, id) => {
      qc.invalidateQueries({ queryKey: QUERY_KEYS.list() });
      qc.removeQueries({ queryKey: QUERY_KEYS.detail(id) });
    },
  });
};

export const guitarQueryKeys = QUERY_KEYS;
