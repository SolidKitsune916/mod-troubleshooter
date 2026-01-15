import { useQuota, getQuotaPercentage, getQuotaStatus } from '@hooks/index.ts';
import './QuotaIndicator.css';

/**
 * Displays the current Nexus API quota status.
 * Shows hourly and daily remaining requests with visual indicators.
 */
export const QuotaIndicator: React.FC = () => {
  const { data: quota, isLoading, error } = useQuota();

  if (isLoading) {
    return (
      <div className="quota-indicator quota-indicator--loading" aria-label="Loading quota information">
        <span className="quota-indicator__label">Quota</span>
        <span className="quota-indicator__value">...</span>
      </div>
    );
  }

  if (error || !quota || !quota.available) {
    return (
      <div
        className="quota-indicator quota-indicator--unavailable"
        aria-label="Quota information unavailable"
        title="Make an API request to see quota information"
      >
        <span className="quota-indicator__label">Quota</span>
        <span className="quota-indicator__value">N/A</span>
      </div>
    );
  }

  const hourlyPercent = getQuotaPercentage(quota, 'hourly');
  const dailyPercent = getQuotaPercentage(quota, 'daily');
  const hourlyStatus = getQuotaStatus(hourlyPercent);
  const dailyStatus = getQuotaStatus(dailyPercent);

  // Use the more restrictive status
  const overallStatus = hourlyStatus === 'critical' || dailyStatus === 'critical'
    ? 'critical'
    : hourlyStatus === 'warning' || dailyStatus === 'warning'
      ? 'warning'
      : 'good';

  return (
    <div
      className={`quota-indicator quota-indicator--${overallStatus}`}
      role="status"
      aria-label={`API quota: ${quota.hourlyRemaining} of ${quota.hourlyLimit} hourly, ${quota.dailyRemaining} of ${quota.dailyLimit} daily`}
    >
      <span className="quota-indicator__label">Quota</span>
      <div className="quota-indicator__meters">
        <div
          className={`quota-indicator__meter quota-indicator__meter--${hourlyStatus}`}
          title={`Hourly: ${quota.hourlyRemaining}/${quota.hourlyLimit} (${hourlyPercent}%)`}
        >
          <span className="quota-indicator__meter-label">H</span>
          <div className="quota-indicator__meter-bar">
            <div
              className="quota-indicator__meter-fill"
              style={{ width: `${hourlyPercent}%` }}
            />
          </div>
          <span className="quota-indicator__meter-value">{quota.hourlyRemaining}</span>
        </div>
        <div
          className={`quota-indicator__meter quota-indicator__meter--${dailyStatus}`}
          title={`Daily: ${quota.dailyRemaining}/${quota.dailyLimit} (${dailyPercent}%)`}
        >
          <span className="quota-indicator__meter-label">D</span>
          <div className="quota-indicator__meter-bar">
            <div
              className="quota-indicator__meter-fill"
              style={{ width: `${dailyPercent}%` }}
            />
          </div>
          <span className="quota-indicator__meter-value">{quota.dailyRemaining}</span>
        </div>
      </div>
    </div>
  );
};

export default QuotaIndicator;
