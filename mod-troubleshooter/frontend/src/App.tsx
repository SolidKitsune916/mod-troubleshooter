/** Main application component */
function App() {
  return (
    <main className="min-h-screen p-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold text-text-primary">
          Mod Troubleshooter
        </h1>
        <p className="mt-2 text-text-secondary">
          Visualize, analyze, and troubleshoot Skyrim SE mod collections
        </p>
      </header>

      <section
        aria-labelledby="getting-started"
        className="rounded-sm border border-border bg-bg-card p-6"
      >
        <h2 id="getting-started" className="mb-4 text-xl font-semibold">
          Getting Started
        </h2>
        <p className="text-text-secondary">
          Enter a Nexus Mods collection URL or slug to begin analyzing your mod
          setup.
        </p>
      </section>
    </main>
  );
}

export default App;
