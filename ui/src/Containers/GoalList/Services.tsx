import restClient from '../../Helpers/RestClient';

export const getGoals = (archived: boolean) =>
  restClient.get(`/api/v1/goals?showArchived=${archived ? 'true' : 'false'}`);
