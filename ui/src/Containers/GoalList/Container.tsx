import React from 'react';

import EmptyBox from '../../Components/Variants/EmptyBox';
import LinkButton from '../../Components/Variants/LinkButton';

const NoGoals = () => (
  <EmptyBox>
    <p>You haven’t added any goals yet.</p>
    <LinkButton href='/create-goal' to='/create-goal'>
      Add a Goal
    </LinkButton>
  </EmptyBox>
);

interface GoalListProps {}

const GoalList = (props: GoalListProps) => {
  return <NoGoals />;
};

export default GoalList;
