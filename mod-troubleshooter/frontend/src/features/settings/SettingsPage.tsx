import { useState, useId, useRef } from 'react';

import {
  useSettings,
  useUpdateSettings,
  useValidateApiKey,
} from '@hooks/useSettings.ts';
import { ApiError } from '@services/api.ts';

/** Props for SettingsPage */
interface SettingsPageProps {
  onBack: () => void;
}

/** Loading skeleton for settings */
const SettingsSkeleton: React.FC = () => (
  <div className="space-y-6 animate-pulse">
    <div className="p-6 rounded-sm bg-bg-card border border-border">
      <div className="h-6 w-1/3 bg-bg-secondary rounded-xs mb-4" />
      <div className="h-4 w-2/3 bg-bg-secondary rounded-xs mb-6" />
      <div className="h-10 w-full bg-bg-secondary rounded-xs mb-4" />
      <div className="h-10 w-32 bg-bg-secondary rounded-xs" />
    </div>
  </div>
);

/** Settings page with API key form */
export const SettingsPage: React.FC<SettingsPageProps> = ({ onBack }) => {
  const apiKeyInputId = useId();
  const errorMessageId = useId();
  const successMessageId = useId();
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: settings, isLoading, error: fetchError, refetch } = useSettings();
  const updateSettings = useUpdateSettings();
  const validateApiKey = useValidateApiKey();

  const [apiKey, setApiKey] = useState('');
  const [showKey, setShowKey] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const handleValidate = async () => {
    setLocalError(null);
    setSuccessMessage(null);

    if (!apiKey.trim()) {
      setLocalError('Please enter an API key to validate');
      inputRef.current?.focus();
      return;
    }

    try {
      const result = await validateApiKey.mutateAsync(apiKey.trim());
      if (result.valid) {
        setSuccessMessage('API key is valid!');
      } else {
        setLocalError('API key is invalid. Please check and try again.');
      }
    } catch (e) {
      if (e instanceof ApiError) {
        setLocalError(e.message);
      } else {
        setLocalError('Failed to validate API key');
      }
    }
  };

  const handleSave = async () => {
    setLocalError(null);
    setSuccessMessage(null);

    try {
      await updateSettings.mutateAsync(apiKey.trim());
      setSuccessMessage('Settings saved successfully!');
      setApiKey(''); // Clear form after save
    } catch (e) {
      if (e instanceof ApiError) {
        setLocalError(e.message);
      } else {
        setLocalError('Failed to save settings');
      }
    }
  };

  const handleClear = async () => {
    setLocalError(null);
    setSuccessMessage(null);

    try {
      await updateSettings.mutateAsync('');
      setSuccessMessage('API key cleared');
      setApiKey('');
    } catch (e) {
      if (e instanceof ApiError) {
        setLocalError(e.message);
      } else {
        setLocalError('Failed to clear API key');
      }
    }
  };

  const isProcessing =
    updateSettings.isPending || validateApiKey.isPending;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <button
          onClick={onBack}
          className="flex items-center gap-2 text-text-secondary hover:text-text-primary
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            transition-colors"
        >
          <span aria-hidden="true">←</span>
          <span>Back to Collections</span>
        </button>
        <h2 className="text-2xl font-bold text-text-primary">Settings</h2>
        <SettingsSkeleton />
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className="space-y-6">
        <button
          onClick={onBack}
          className="flex items-center gap-2 text-text-secondary hover:text-text-primary
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            transition-colors"
        >
          <span aria-hidden="true">←</span>
          <span>Back to Collections</span>
        </button>
        <h2 className="text-2xl font-bold text-text-primary">Settings</h2>
        <div
          role="alert"
          className="p-6 rounded-sm bg-error/10 border border-error text-center"
        >
          <p className="text-error font-medium mb-4">
            Failed to load settings
          </p>
          <button
            onClick={() => refetch()}
            className="min-h-11 px-6 py-2 rounded-sm
              bg-error text-white font-medium
              hover:bg-error/80
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <button
        onClick={onBack}
        className="flex items-center gap-2 text-text-secondary hover:text-text-primary
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors"
      >
        <span aria-hidden="true">←</span>
        <span>Back to Collections</span>
      </button>

      <h2 className="text-2xl font-bold text-text-primary">Settings</h2>

      <div className="p-6 rounded-sm bg-bg-card border border-border">
        <h3 className="text-lg font-semibold text-text-primary mb-2">
          Nexus Mods API Key
        </h3>
        <p className="text-text-secondary text-sm mb-4">
          Enter your Nexus Mods API key to enable collection browsing. You can
          find your API key in your{' '}
          <a
            href="https://www.nexusmods.com/users/myaccount?tab=api"
            target="_blank"
            rel="noopener noreferrer"
            className="text-accent hover:underline focus-visible:outline-3 focus-visible:outline-focus"
          >
            Nexus Mods account settings
          </a>
          .
        </p>

        {/* Current status */}
        <div className="mb-4 p-3 rounded-xs bg-bg-secondary">
          <p className="text-sm text-text-secondary">
            <span className="font-medium">Current status:</span>{' '}
            {settings?.keyConfigured ? (
              <span className="text-success">
                API key configured ({settings.nexusApiKey})
              </span>
            ) : (
              <span className="text-warning">No API key configured</span>
            )}
          </p>
        </div>

        {/* Success message */}
        {successMessage && (
          <div
            id={successMessageId}
            role="status"
            aria-live="polite"
            className="mb-4 p-3 rounded-xs bg-success/10 border border-success"
          >
            <p className="text-success text-sm">{successMessage}</p>
          </div>
        )}

        {/* Error message */}
        {localError && (
          <div
            id={errorMessageId}
            role="alert"
            className="mb-4 p-3 rounded-xs bg-error/10 border border-error"
          >
            <p className="text-error text-sm">{localError}</p>
          </div>
        )}

        {/* API Key input */}
        <div className="mb-4">
          <label
            htmlFor={apiKeyInputId}
            className="block text-sm font-medium text-text-primary mb-2"
          >
            API Key {!settings?.keyConfigured && <span className="text-error">*</span>}
          </label>
          <div className="relative">
            <input
              ref={inputRef}
              id={apiKeyInputId}
              type={showKey ? 'text' : 'password'}
              value={apiKey}
              onChange={(e) => {
                setApiKey(e.target.value);
                setLocalError(null);
                setSuccessMessage(null);
              }}
              placeholder={
                settings?.keyConfigured
                  ? 'Enter new API key to update'
                  : 'Enter your Nexus Mods API key'
              }
              autoComplete="off"
              aria-required={!settings?.keyConfigured}
              aria-invalid={!!localError}
              aria-describedby={localError ? errorMessageId : undefined}
              disabled={isProcessing}
              className="w-full min-h-11 px-4 py-2 pr-24 rounded-sm
                bg-bg-secondary border border-border text-text-primary
                placeholder:text-text-muted
                hover:border-border-hover
                focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/20
                disabled:opacity-50 disabled:cursor-not-allowed
                transition-colors"
            />
            <button
              type="button"
              onClick={() => setShowKey(!showKey)}
              disabled={isProcessing}
              className="absolute right-2 top-1/2 -translate-y-1/2
                px-3 py-1 text-xs font-medium text-text-secondary
                hover:text-text-primary
                focus-visible:outline-3 focus-visible:outline-focus
                disabled:opacity-50
                transition-colors"
              aria-label={showKey ? 'Hide API key' : 'Show API key'}
            >
              {showKey ? 'Hide' : 'Show'}
            </button>
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex flex-wrap gap-3">
          <button
            onClick={handleValidate}
            disabled={isProcessing || !apiKey.trim()}
            className="min-h-11 px-6 py-2 rounded-sm
              bg-bg-secondary border border-border text-text-primary font-medium
              hover:bg-bg-tertiary hover:border-border-hover
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              disabled:opacity-50 disabled:cursor-not-allowed
              transition-colors motion-reduce:transition-none"
          >
            {validateApiKey.isPending ? 'Validating...' : 'Validate'}
          </button>

          <button
            onClick={handleSave}
            disabled={isProcessing || !apiKey.trim()}
            className="min-h-11 px-6 py-2 rounded-sm
              bg-accent text-white font-medium
              hover:bg-accent/80
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              disabled:opacity-50 disabled:cursor-not-allowed
              transition-colors motion-reduce:transition-none"
          >
            {updateSettings.isPending ? 'Saving...' : 'Save'}
          </button>

          {settings?.keyConfigured && (
            <button
              onClick={handleClear}
              disabled={isProcessing}
              className="min-h-11 px-6 py-2 rounded-sm
                bg-error/10 border border-error text-error font-medium
                hover:bg-error/20
                focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                disabled:opacity-50 disabled:cursor-not-allowed
                transition-colors motion-reduce:transition-none"
            >
              Clear Key
            </button>
          )}
        </div>
      </div>

      {/* Help section */}
      <div className="p-6 rounded-sm bg-bg-card border border-border">
        <h3 className="text-lg font-semibold text-text-primary mb-2">
          About API Keys
        </h3>
        <ul className="text-text-secondary text-sm space-y-2 list-disc list-inside">
          <li>
            A Nexus Mods API key is required to browse collections
          </li>
          <li>
            Free accounts have limited API requests (1000/day)
          </li>
          <li>
            Premium accounts have higher limits (2500/day)
          </li>
          <li>
            Your API key is stored locally and sent directly to Nexus Mods
          </li>
        </ul>
      </div>
    </div>
  );
};

export default SettingsPage;
