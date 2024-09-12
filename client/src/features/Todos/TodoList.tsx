import { List, ThemeIcon } from "@mantine/core";
import { CheckCircleFillIcon } from "@primer/octicons-react";
import { ENDPOINT } from "../../utils/fetcher";
import { KeyedMutator } from "swr";
import { Todo } from "./types";

type TodoListProps = {
  data?: Todo[];
  mutate: KeyedMutator<Todo[]>;
};
export const TodoList: React.FC<TodoListProps> = ({ data, mutate }) => {
  const handleStatusUpdate = async (id: number) => {
    const updated = await fetch(`${ENDPOINT}/api/todos/${id}/status`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
    }).then((r) => {
      return r.json();
    });

    await mutate(updated);
  };

  if (!data?.length) {
    return <div>No todos</div>;
  }

  return (
    <List spacing={"xs"} size="sm" mb={12} center>
      {data?.map((todo) => {
        return (
          <List.Item
            key={`todo__${todo.id}`}
            icon={
              todo.done ? (
                <ThemeIcon
                  color="teal"
                  size={24}
                  radius="xl"
                  onClick={() => handleStatusUpdate(todo.id)}
                >
                  <CheckCircleFillIcon size={20} />
                </ThemeIcon>
              ) : (
                <ThemeIcon
                  color="gray"
                  size={24}
                  radius="xl"
                  onClick={() => handleStatusUpdate(todo.id)}
                >
                  <CheckCircleFillIcon />
                </ThemeIcon>
              )
            }
          >
            {todo.title}
          </List.Item>
        );
      })}
    </List>
  );
};
