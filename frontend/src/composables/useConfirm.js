export function useConfirm() {
  function confirm(message) {
    return window.confirm(message)
  }
  return { confirm }
}
