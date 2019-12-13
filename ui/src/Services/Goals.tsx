import restClient from '../Helpers/RestClient';

const getNormalizedDate = () => {
  const now = new Date();
  return new Date(now.getTime() - now.getTimezoneOffset() * 60 * 1000);
};

export const getGoals = (archived?: boolean): Promise<{ data: Array<IGoal> }> =>
  restClient.get(
    `/api/v1/goals?showArchived=${archived === true ? 'true' : 'false'}`
  );

export const getGoal = (goalId: string): Promise<{ data: IGoal }> =>
  restClient.get(`/api/v1/goals?${goalId}`);

export const getGoalData = (goalId: string): Promise<{ data: IGoalData }> =>
  restClient.get(`/api/v1/goals?${goalId}/data`);

export const getGoalETA = (goalId: string): Promise<{ data: IGoalData }> =>
  restClient.get(`/api/v1/goals?${goalId}/eta`);

export const getGoalRawData = (goalId: string): Promise<{ data: IGoalData }> =>
  restClient.get(`/api/v1/goals?${goalId}/raw-data`);

export const addGoalData = (goalId: string, data: IGoalData) =>
  restClient.post(`/api/v1/goals?${goalId}/data`, data);

export const addGoalDataPoint = (goalId: string, value: number) =>
  restClient.post(`/api/v1/goals?${goalId}/data/single?add=true`, {
    value,
    normalized: getNormalizedDate(),
  });

export const addGoalDataSetPoint = (goalId: string, value: number) =>
  restClient.post(`/api/v1/goals?${goalId}/data/single?add=false`, {
    value,
    normalized: getNormalizedDate(),
  });

export const createGoal = (goal: IGoal) =>
  restClient.post('/api/v1/goals', goal);

export const updateGoal = (goal: IGoal) =>
  restClient.put(`/api/v1/goals/${goal.id}`, goal);

export const deleteGoal = (goalId: string) =>
  restClient.delete('/api/v1/goals/${goalId}');

export const unarchiveGoal = (goal: IGoal) =>
  updateGoal({ ...goal, archived: false });
