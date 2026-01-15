import { useState, useId, useRef } from 'react';

import {
  useSettings,
  useUpdateSettings,
  useValidateApiKey,
} from '@hooks/useSettings.ts';
import { ApiError } from '@services/api.ts';
import styles from './SettingsPage.module.css';

/** Props for SettingsPage */
interface SettingsPageProps {
  onBack: () => void;
}

/** Loading skeleton for settings */
const SettingsSkeleton: React.FC = () => (
  <div className={styles.skeleton}>
    <div className={styles.skeletonCard}>
      <div className={styles.skeletonLine} style={{ width: '33%', marginBottom: 'var(--spacing-4)' }} />
      <div className={styles.skeletonLine} style={{ width: '66%', marginBottom: 'var(--spacing-6)' }} />
      <div className={styles.skeletonLine} style={{ width: '100%', marginBottom: 'var(--spacing-4)', height: '44px' }} />
      <div className={styles.skeletonLine} style={{ width: '128px', height: '44px' }} />
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

  const isProcessing = updateSettings.isPending || validateApiKey.isPending;

  if (isLoading) {
    return (
      <div className={styles.container}>
        <button onClick={onBack} className={styles.backButton}>
          <span aria-hidden="true">←</span>
          <span>Back to Collections</span>
        </button>
        <h2 className={styles.pageTitle}>Settings</h2>
        <SettingsSkeleton />
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className={styles.container}>
        <button onClick={onBack} className={styles.backButton}>
          <span aria-hidden="true">←</span>
          <span>Back to Collections</span>
        </button>
        <h2 className={styles.pageTitle}>Settings</h2>
        <div role="alert" className={styles.errorMessage}>
          <p className={styles.errorText}>Failed to load settings</p>
        </div>
        <button onClick={() => refetch()} className={styles.btnDanger}>
          Try Again
        </button>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <button onClick={onBack} className={styles.backButton}>
        <span aria-hidden="true">←</span>
        <span>Back to Collections</span>
      </button>

      <h2 className={styles.pageTitle}>Settings</h2>

      <div className={styles.card}>
        <h3 className={styles.cardTitle}>Nexus Mods API Key</h3>
        <p className={styles.cardDescription}>
          Enter your Nexus Mods API key to enable collection browsing. You can
          find your API key in your{' '}
          <a
            href="https://www.nexusmods.com/users/myaccount?tab=api"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.link}
          >
            Nexus Mods account settings
          </a>
          .
        </p>

        {/* Current status */}
        <div className={styles.statusBox}>
          <p className={styles.statusText}>
            <span className={styles.statusLabel}>Current status:</span>{' '}
            {settings?.keyConfigured ? (
              <span className={styles.statusConfigured}>
                API key configured ({settings.nexusApiKey})
              </span>
            ) : (
              <span className={styles.statusNotConfigured}>No API key configured</span>
            )}
          </p>
        </div>

        {/* Success message */}
        {successMessage && (
          <div
            id={successMessageId}
            role="status"
            aria-live="polite"
            className={styles.successMessage}
          >
            <p className={styles.successText}>{successMessage}</p>
          </div>
        )}

        {/* Error message */}
        {localError && (
          <div
            id={errorMessageId}
            role="alert"
            className={styles.errorMessage}
          >
            <p className={styles.errorText}>{localError}</p>
          </div>
        )}

        {/* API Key input */}
        <div className={styles.formGroup}>
          <label htmlFor={apiKeyInputId} className={styles.label}>
            API Key {!settings?.keyConfigured && <span className={styles.required}>*</span>}
          </label>
          <div className={styles.inputWrapper}>
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
              className={styles.input}
            />
            <button
              type="button"
              onClick={() => setShowKey(!showKey)}
              disabled={isProcessing}
              className={styles.showButton}
              aria-label={showKey ? 'Hide API key' : 'Show API key'}
            >
              {showKey ? 'Hide' : 'Show'}
            </button>
          </div>
        </div>

        {/* Action buttons */}
        <div className={styles.buttonGroup}>
          <button
            onClick={handleValidate}
            disabled={isProcessing || !apiKey.trim()}
            className={styles.btnSecondary}
          >
            {validateApiKey.isPending ? 'Validating...' : 'Validate'}
          </button>

          <button
            onClick={handleSave}
            disabled={isProcessing || !apiKey.trim()}
            className={styles.btnPrimary}
          >
            {updateSettings.isPending ? 'Saving...' : 'Save'}
          </button>

          {settings?.keyConfigured && (
            <button
              onClick={handleClear}
              disabled={isProcessing}
              className={styles.btnDanger}
            >
              Clear Key
            </button>
          )}
        </div>
      </div>

      {/* Help section */}
      <div className={styles.card}>
        <h3 className={styles.cardTitle}>About API Keys</h3>
        <ul className={styles.helpList}>
          <li>A Nexus Mods API key is required to browse collections</li>
          <li>Free accounts have limited API requests (1000/day)</li>
          <li>Premium accounts have higher limits (2500/day)</li>
          <li>Your API key is stored locally and sent directly to Nexus Mods</li>
        </ul>
      </div>
    </div>
  );
};

export default SettingsPage;
