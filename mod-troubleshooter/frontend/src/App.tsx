import { CollectionBrowser } from '@features/collections/index.ts';

/** Main application component */
function App() {
  return (
    <main className="min-h-screen p-8 max-w-6xl mx-auto">
      <header className="mb-8">
        <h1 className="text-3xl font-bold text-text-primary">
          Mod Troubleshooter
        </h1>
        <p className="mt-2 text-text-secondary">
          Visualize, analyze, and troubleshoot Skyrim SE mod collections
        </p>
      </header>

      <CollectionBrowser />
    </main>
  );
}

export default App;
