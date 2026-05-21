import { useEffect, useState } from 'react';

interface PictureGalleryProps {
  pictures: string[];
}

export const PictureGallery = ({ pictures }: PictureGalleryProps) => {
  const [preview, setPreview] = useState<string | null>(null);

  useEffect(() => {
    if (!preview) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setPreview(null);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [preview]);

  if (pictures.length === 0) {
    return <p className="text-sm text-slate-500">No pictures yet.</p>;
  }

  return (
    <>
      <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
        {pictures.map((src, index) => (
          <li key={`${src}-${index}`} className="aspect-square overflow-hidden rounded-md border border-slate-200 bg-white">
            <button
              type="button"
              onClick={() => setPreview(src)}
              className="block h-full w-full focus:outline-none focus:ring-2 focus:ring-slate-500"
              aria-label={`Preview picture ${index + 1}`}
            >
              <img
                src={src}
                alt={`Guitar picture ${index + 1}`}
                className="h-full w-full object-cover"
                loading="lazy"
              />
            </button>
          </li>
        ))}
      </ul>

      {preview ? (
        <div
          role="dialog"
          aria-modal="true"
          aria-label="Picture preview"
          onClick={() => setPreview(null)}
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-6"
        >
          <img
            src={preview}
            alt="Preview"
            className="max-h-full max-w-full rounded-md shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          />
          <button
            type="button"
            onClick={() => setPreview(null)}
            className="absolute right-4 top-4 rounded-md bg-white/90 px-3 py-1 text-sm font-medium text-slate-900 hover:bg-white"
          >
            Close
          </button>
        </div>
      ) : null}
    </>
  );
};
