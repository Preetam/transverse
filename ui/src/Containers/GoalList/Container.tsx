import React from 'react';
import goalServices from './Services';

import EmptyBox from '../../StyledComponents/Variants/Box';
import LinkButton from '../../StyledComponents/Variants/LinkButton';
import { H1 } from '../../StyledComponents/Basic/Header';

const NoGoals = () => (
  <EmptyBox>
    <p>You haven’t added any goals yet.</p>
    <LinkButton href="/create-goal" to="/create-goal">
      Add a Goal
    </LinkButton>
  </EmptyBox>
);

interface GoalListProps {}

const GoalList = (props: GoalListProps) => {
  const [goalsList, setGoalsList] = React.useState<Array<IGoal>>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  React.useEffect(() => {
    goalServices
      .getGoals()
      .then(({ data }: { data: Array<IGoal> }) => {
        setGoalsList(data);
        setError(false);
      })
      .catch((ex: Error) => {
        setError(true);
        console.error(ex);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);
  return (
    <EmptyBox flexDirection="column">
      <EmptyBox>
        <H1>Goals</H1>
      </EmptyBox>
      <EmptyBox flexDirection="column">
        {loading === false && goalsList.length === 0 && <NoGoals />}
        {loading === true && 'loading'}
        {loading === false && goalsList.length > 0}
      </EmptyBox>
    </EmptyBox>
  );
};

export default GoalList;
