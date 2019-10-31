interface IGoal {
  id: string;
  name: string;
  user: string;
  description: string;
  target: number;
  archived: boolean;
  created: number;
  updated: number;
  deleted: number;
}

interface DataPoint {
  ts: number;
}

interface IGoalData {
  series: Array<DataPoint>;
  prediction: Array<DataPoint>;
  eta: number;
}
