import { useCallback } from 'react';
import { useHistory } from 'react-router-dom';

export default function useGoToPath(path: string) {
  const history = useHistory();
  return useCallback(() => history.push(path), [history, path]);
}
