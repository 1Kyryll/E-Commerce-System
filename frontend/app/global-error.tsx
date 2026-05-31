"use client";
export default function GlobalError({ reset }: { reset: () => void }) {
  return (
    <html>
      <body className="min-h-screen flex items-center justify-center p-8">
        <div className="text-center">
          <h1 className="text-2xl mb-2">Something broke.</h1>
          <button onClick={() => reset()} className="underline">Try again</button>
        </div>
      </body>
    </html>
  );
}
