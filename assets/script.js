/* placeholder file for JavaScript */
const confirm_delete = (id) => {
  if(window.alert(`Task ${id} を削除します．よろしいですか？`)) {
      location.href = `/task/delete/${id}`;
  }
}
const confirm_update = (id) => {
  if(window.alert(`Task ${id} を更新します．よろしいですか？`)) {
      location.href = `/task/edit/${id}`;
  }
}
